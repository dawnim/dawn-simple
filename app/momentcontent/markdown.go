// Package momentcontent identifies media in the community Markdown dialect.
package momentcontent

import (
	"regexp"
	"slices"
	"strconv"

	sdk "github.com/dawnim/dawn-server/app/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gorm.io/gorm"
)

var attachmentTag = regexp.MustCompile("^\\[attachment://([1-9][0-9]*)\\]")
var htmlMedia = regexp.MustCompile("(?i)<(img|video|audio|iframe)\\b")
var mediaMarker = regexp.MustCompile("(?i)MEDIA::(audio|video|iframe)::https?://")

var markdown = goldmark.New(goldmark.WithParserOptions(
	parser.WithInlineParsers(util.Prioritized(attachmentParser{}, 199)),
))

var attachmentKind = ast.NewNodeKind("MomentAttachment")

type attachment struct {
	ast.BaseInline
	id uint
}

func (n *attachment) Kind() ast.NodeKind { return attachmentKind }
func (n *attachment) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

type attachmentParser struct{}

func (attachmentParser) Trigger() []byte { return []byte{'['} }
func (attachmentParser) Parse(_ ast.Node, reader text.Reader, _ parser.Context) ast.Node {
	line, position := reader.PeekLine()
	match := attachmentTag.FindSubmatchIndex(line)
	if match == nil || (len(line) > match[1] && (line[match[1]] == '(' || line[match[1]] == '[')) || (position.Start > 0 && reader.Source()[position.Start-1] == '!') {
		return nil
	}
	id, err := strconv.ParseUint(string(line[match[2]:match[3]]), 10, strconv.IntSize)
	if err != nil {
		return nil
	}
	reader.Advance(match[1])
	return &attachment{id: uint(id)}
}

type Media struct {
	IDs    []uint
	Inline bool
}

// ParseMedia uses Markdown nodes so escaped references and code stay literal.
func ParseMedia(content string) Media {
	source := []byte(content)
	media := Media{}
	_ = ast.Walk(markdown.Parser().Parse(text.NewReader(source)), func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := node.(type) {
		case *attachment:
			if !slices.Contains(media.IDs, n.id) {
				media.IDs = append(media.IDs, n.id)
			}
			media.Inline = true
		case *ast.Image:
			media.Inline = true
		case *ast.RawHTML:
			media.Inline = media.Inline || htmlMedia.Match(segmentValue(n.Segments, source))
		case *ast.HTMLBlock:
			media.Inline = media.Inline || htmlMedia.Match(segmentValue(n.Lines(), source))
		case *ast.Text:
			media.Inline = media.Inline || mediaMarker.Match(n.Segment.Value(source))
		case *ast.CodeSpan:
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return media
}

func segmentValue(segments *text.Segments, source []byte) []byte {
	var value []byte
	for i := 0; i < segments.Len(); i++ {
		segment := segments.At(i)
		value = append(value, segment.Value(source)...)
	}
	return value
}

var ImageMIMEs = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

// Images scopes community references to supported attachment types.
func Images(db *gorm.DB, ids []uint) *gorm.DB {
	return db.Model(&sdk.Attachment{}).Where("id IN ?", ids).Where(
		"file_id IN (?)", db.Model(&sdk.File{}).Select("id").Where("mime IN ?", ImageMIMEs),
	)
}
