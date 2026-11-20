(() => {
    "use strict";
    const $ = (id) => document.getElementById(id);
    let momentEditor;
    const markdownReady = window.momentMarkdownReady || Promise.resolve();
    const element = (tag, className, value) => {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (value !== undefined) node.textContent = value;
        return node;
    };
    let language = localStorage.getItem("dawn.site.locale") || (navigator.language.startsWith("zh") ? "zh" : "en");
    const english = {
        黎明: "Dawn", 社区首页: "Home", 搜索: "Search", "搜索主题和内容": "Search topics and posts", 交流板块: "Topics", 灵感在这里相遇: "Where ideas meet", 新动态: "New posts", 新评论: "New replies", 热门贴: "Hot posts", 我发的: "Mine", 全部: "All", 自由交流: "Discussion", 问题求助: "Help", 作品分享: "Showcase", AI趣闻: "AI stories", 社区公约: "Community guidelines", "分享经验，认真提问，友善交流。鲜花送给灵感，鸡蛋留给玩笑": "Share what you learn, ask thoughtful questions, and be kind. Flowers for inspiration; eggs for a little fun", 认识黎明: "Meet Dawn", 本周热议: "Trending this week", 社区足迹: "Community", 主题: "Topics", 成员: "Members", "让每一个想法，都有回响": "Give every idea a voice", "欢迎来到黎明社区": "Welcome to Dawn Community", "登录后参与交流": "Sign in to join the conversation", 标题: "Title", 内容: "Content", "给你的想法起个标题": "Give your idea a title", 添加图片: "Add images", "每天最多发布2条，回复不限": "Up to 2 posts a day, unlimited replies", 暂无热议: "No trending topics yet", 复制链接: "Copy link", 复制: "Copy", 已复制: "Copied", 编辑: "Edit", 草稿已保存: "Draft saved", 发消息: "Message", "期待你的第一条分享": "Be the first to share an idea", "没有找到相关主题": "No matching topics", "清除筛选": "Clear filters", "楼主": "Author", "更早的消息": "Earlier messages", "请登录后查看自己的动态": "Sign in to see your posts",
        使用IP属地: "Use IP region", "发布次数按站点设置，回复不限": "Daily post limit applies; replies are unlimited", "登录后参与聊天": "Sign in to send messages", "暂无消息，来打个招呼吧": "No messages yet. Say hello", 新消息: "New messages",
        编辑动态: "Edit post", 保存修改: "Save changes", 图库图片: "Gallery images", 上传图片: "Upload images", 搜索图片: "Search images", 移除图片: "Remove image", 正在上传图片: "Uploading images", 暂无图片: "No images",
        网友动态: "Community", "分享此刻，发现有趣的人和事": "Share a moment and discover the community", 发布动态: "New post", 分享你的想法: "Share your thoughts", "添加地点（选填）": "Add a location (optional)", 发布: "Post", 暂无动态: "No posts yet", 评论: "Comments", 回复: "Reply", 删除: "Delete", 发送: "Send", 写下你的评论: "Write a comment", 上一页: "Previous", 下一页: "Next", 返回动态: "All posts", 未知: "Unknown", "请先登录或完善信息": "Please sign in or complete your profile first", "确认删除？": "Delete this?", 鲜花: "Flowers", 鸡蛋: "Eggs", 加载中: "Loading", "AI写动态": "Create with AI", "正在生成": "Generating", "添加图片（最多9张）": "Add images (up to 9)", "最多选择9张图片": "Choose up to 9 images",
        动态: "Moment", 帖子: "Post", 匿名发布: "Post anonymously", 匿名用户: "Anonymous user", 正文: "Write", 实时预览: "Live preview", 预览: "Preview", 进入全屏: "Fullscreen", 退出全屏: "Exit fullscreen", 一级标题: "Heading 1", 二级标题: "Heading 2", 三级标题: "Heading 3", 行内代码: "Inline code", 代码块: "Code block", 无序列表: "Bullet list", 星号列表: "Star list", 有序列表: "Numbered list", 待办列表: "Task list", 左对齐: "Align left", 居中: "Center", 右对齐: "Align right", 插入链接: "Insert link", 链接文字: "Link text", 链接地址: "Link URL", 插入表格: "Insert table", 选择表格行列: "Choose table size", 插入媒体: "Insert media", 填写音频或视频地址: "Enter an audio or video URL", 媒体地址: "Media URL", 插入剧透: "Insert spoiler", 选择剧透样式: "Choose a spoiler style", 可折叠剧透: "Collapsible spoiler", 可折叠剧透说明: "Insert a collapsible spoiler block", 点击遮罩: "Click to reveal", 点击遮罩说明: "Hide the text until it is clicked", 选择表情: "Choose an emoji", Markdown帮助: "Markdown help", 分割线: "Divider", 取消: "Cancel", 插入: "Insert", 关闭: "Close",
        跟随系统: "System theme",
        亮色: "Light theme",
        暗色: "Dark theme",
        登录: "Sign in",
        退出登录: "Sign out",
        "从想法到产品，简单一点": "A simpler path from idea to product",
        "开放的CMS框架，让管理、实时通信与AI客服自然融入你的业务":
            "An open CMS framework that brings administration, realtime communication, and AI support into your product",
        开始构建: "Start building",
        体验在线客服: "Try customer support",
        "一套清晰的扩展能力，连接你的网站和用户":
            "Clear extension points. Connected websites and customers",
        专注你的业务: "Focus on your product",
        "用户、权限、文件与实时通信已经就绪，按需扩展，无需修改框架源码":
            "Users, permissions, files, and realtime are ready. Extend what you need through public APIs",
        让知识参与服务: "Put knowledge to work",
        "文档与业务工具共同支持AI回答，适用条件、身份验证和引用都有清晰边界":
            "Ground AI answers in documentation and business tools, with explicit identity and source boundaries",
        随时转交给人: "Bring a person into the conversation",
        "从网站悬浮入口到Telegram，把AI与人工接待放进同一个会话流程":
            "Connect website widgets and Telegram to one AI and human support flow",
        "从文档开始，也可以直接聊聊": "Read the docs, or start a conversation",
        "客服演示使用Dawn与PrismGo文档，支持AI接待和转人工":
            "Try support grounded in Dawn and PrismGo documentation, with AI and human assistance",
        "简洁、开放、可扩展": "Simple. Open. Extensible",
        用户名: "Username",
        密码: "Password",
        昵称: "Nickname",
        注册: "Create account",
        完善信息: "Complete profile",
        游客: "Guest",
        "还没有账号，去注册": "Create a new account",
        "已有账号，去登录": "Already have an account? Sign in",
        请求失败: "Request failed",
        关闭: "Close",
        聊天: "Chat", 聊天大厅: "Community hall", 输入消息: "Write a message", 查看动态: "View post", 查看评论: "View comment", 取消回复: "Cancel reply", 评论不存在或已删除: "This comment no longer exists",
        "Dawn 文档": "Dawn documentation",
    };
    const icons = {
        search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 4.5 4.5"/>',
        home: '<path d="m3 10 9-7 9 7v11h-6v-7H9v7H3Z"/>',
        chat: '<path d="M4 4h16v12H9l-5 4Z"/><path d="M8 8h8M8 12h5"/>',
        help: '<circle cx="12" cy="12" r="9"/><path d="M9.5 9a2.5 2.5 0 0 1 5 0c0 2-2.5 2-2.5 4m0 3h.01"/>',
        grid: '<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>',
        spark: '<path d="m12 3 2.5 6.5L21 12l-6.5 2.5L12 21l-2.5-6.5L3 12l6.5-2.5Z"/>',
        fire: '<path d="M13 3c1 5-5 5-4 10-2-1-2-3-2-3-5 7 0 11 5 11s10-6 6-11c0 3-2 4-2 4 1-6-3-11-3-11Z"/>',
        people: '<circle cx="9" cy="7" r="3"/><path d="M3 21v-3a6 6 0 0 1 12 0v3M16 4a3 3 0 0 1 0 6m2 4a6 6 0 0 1 3 5v2"/>',
        person: '<circle cx="12" cy="7" r="4"/><path d="M4 21v-2a8 8 0 0 1 16 0v2"/>',
        clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
        plus: '<path d="M12 5v14M5 12h14"/>',
        close: '<path d="m6 6 12 12M6 18 18 6"/>',
        image: '<rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8" cy="8" r="1.5"/><path d="m3 17 5-5 4 4 4-6 5 6"/>',
        pin: '<path d="M20 10c0 6-8 11-8 11S4 16 4 10a8 8 0 1 1 16 0Z"/><circle cx="12" cy="10" r="2.5"/>',
        flower: '<path d="M12 14c-6 0-9-7-5-7 0-6 5-6 5-2 3-5 8-1 5 2 5 1 2 7-5 7Zm0 0v8m0-3c-4 0-6-3-6-3 4-1 6 3 6 3Z"/>',
        egg: '<path d="M19 14c0 5-3 7-7 7s-7-2-7-7S9 3 12 3s7 6 7 11Z"/>',
        link: '<path d="m10 13 4-4m-7 6-1 1a4 4 0 0 0 6 6l4-4a4 4 0 0 0 0-6m1-3 1-1a4 4 0 0 0-6-6L8 6a4 4 0 0 0 0 6"/>',
        trash: '<path d="M3 6h18M9 6V3h6v3m-10 0 1 15h12l1-15M10 10v7m4-7v7"/>',
        logout: '<path d="M10 4H4v16h6m5-12 4 4-4 4m-6-4h14"/>',
        edit: '<path d="M4 20h4L20 8l-4-4L4 16Zm10-14 4 4"/>',
    };
    function icon(name) {
        const node = document.createElementNS("http://www.w3.org/2000/svg", "svg");
        node.setAttribute("viewBox", "0 0 24 24"); node.setAttribute("fill", "none"); node.setAttribute("stroke", "currentColor"); node.setAttribute("stroke-width", "1.65"); node.setAttribute("stroke-linecap", "round"); node.setAttribute("stroke-linejoin", "round"); node.setAttribute("aria-hidden", "true"); node.innerHTML = icons[name];
        return node;
    }
    document.querySelectorAll("[data-icon]").forEach(node => node.replaceChildren(icon(node.dataset.icon)));
    const t = (key) => (language === "zh" ? key : english[key] || key);
    const documentation = $("documentation");
    function locale(value) {
        language = value === "en" ? "en" : "zh";
        localStorage.setItem("dawn.site.locale", language);
        document.documentElement.lang = language;
        document.querySelectorAll("[data-t]").forEach((element) => {
            element.textContent = t(element.dataset.t);
        });
        document.querySelectorAll("[data-placeholder]").forEach((element) => { element.placeholder = t(element.dataset.placeholder); });
        document.querySelectorAll("[data-title]").forEach(node => { node.title = t(node.dataset.title); node.ariaLabel = node.title; });
        $("language").value = language;
        $("close").ariaLabel = t("关闭");
        documentation.title = t("Dawn 文档");
        momentEditor?.setOptions({ locale: language === "en" ? "en" : "zh", placeholder: t("分享你的想法") });
        profile();
        authMode();
        theme();
        if (site?.support_key) window.DawnSupport.setLocale(language);
    }
    let documentURL = "",
        homeScroll = 0;
    function navigate(value, action = "push", fromFrame = false) {
        const url = new URL(value, location.href);
        if (
            url.origin !== location.origin ||
            (url.pathname !== "/" && url.pathname !== "/dawn" && !url.pathname.startsWith("/docs/") && !/^\/moments(?:\/\d+)?\/?$/.test(url.pathname))
        )
            return false;
        url.searchParams.delete("__content");
        if (!fromFrame && url.pathname.startsWith("/docs/")) {
            const path = url.pathname.slice(6)
                .replace(/^prismgo\/README_en(?:\.html)?$/, "en/prismgo/")
                .replace(/^prismgo\/README(?:\.html)?$/, "prismgo/")
                .replace(/^prismgo\/en\//, "en/prismgo/")
                .replace(/^prismgo\/zh_CN\//, "prismgo/")
                .replace(/^en\//, "");
            url.pathname = `/docs/${language === "en" ? "en/" : ""}${path}`;
        }
        const docs = url.pathname.startsWith("/docs/");
        const moments = url.pathname === "/" || url.pathname.startsWith("/moments");
        if (docs && documentation.hidden) homeScroll = scrollY;
        if (action !== "none" && url.href !== location.href)
            history[action === "replace" ? "replaceState" : "pushState"](
                null,
                "",
                url,
            );
        document.documentElement.dataset.page = docs ? "docs" : moments ? "moments" : "home";
        $("nav-dawn").setAttribute("aria-current", String(!docs && !moments));
        $("moments").hidden = !moments;
        ++momentVersion;
        if (moments) queueMicrotask(() => loadMoments());
        documentation.hidden = !docs;
        $("moment-ai").hidden = !docs;
        if (docs) {
            if (!fromFrame && documentURL !== url.href) {
                const content = new URL(url);
                content.searchParams.set("__content", "1");
                if (documentation.hasAttribute("src"))
                    documentation.contentWindow.location.replace(content);
                else documentation.src = content;
            }
            documentURL = url.href;
        } else {
            document.title = moments ? `${t("网友动态")} · Dawn` : "Dawn";
            scrollTo(0, moments ? 0 : homeScroll);
            if (site?.support_key) window.DawnSupport.setTheme(mode);
        }
        return true;
    }
    document.addEventListener("click", (event) => {
        const link = event.target.closest("a[href]");
        if (
            !link ||
            event.defaultPrevented ||
            event.button ||
            event.metaKey ||
            event.ctrlKey ||
            event.shiftKey ||
            event.altKey ||
            link.target === "_blank" ||
            link.hasAttribute("download")
        )
            return;
        if (navigate(link.href)) event.preventDefault();
    });
    addEventListener("popstate", () => navigate(location.href, "none"));
    addEventListener("message", (event) => {
        if (
            event.origin !== location.origin ||
            event.source !== documentation.contentWindow ||
            event.data?.source !== "dawn-docs" ||
            documentation.hidden
        )
            return;
        if (event.data.type === "click") {
            if (innerWidth > 600) { window.DawnSupport?.close(); $("community-chat").hidden = true; }
            return;
        }
        if (event.data.type === "navigate")
            navigate(event.data.url, event.data.action, true);
        if (event.data.type === "state") {
            if (event.data.locale && event.data.locale !== language) locale(event.data.locale);
            document.title = event.data.title;
            if (site?.support_key)
                window.DawnSupport.setTheme(event.data.theme);
        }
    });
    let token = localStorage.getItem("dawn.member") || "",
        user,
        site,
        register = false,
        accountVersion = 0;
    const media = matchMedia("(prefers-color-scheme: dark)");
    const modes = ["system", "light", "dark"],
        themes = {
            system: [
                "跟随系统",
                '<rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8m-4-4v4"/>',
            ],
            light: [
                "亮色",
                '<circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5"/>',
            ],
            dark: [
                "暗色",
                '<path d="M20.8 13A9 9 0 0 1 11 3.2 9 9 0 1 0 20.8 13Z"/>',
            ],
        };
    let mode = localStorage.getItem("dawn.site.theme") || "system";
    function theme() {
        document.documentElement.dataset.theme =
            mode === "system" ? (media.matches ? "dark" : "light") : mode;
        const [label, icon] = themes[mode];
        $("theme").title = t(label);
        $("theme").ariaLabel = t(label);
        $("theme").innerHTML =
            `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icon}</svg>`;
        if (site?.support_key) window.DawnSupport.setTheme(mode);
    }
    locale(language);
    const changeLanguage = (event) => {
        if (event.target.value === language) return;
        locale(event.target.value);
        if (!documentation.hidden) navigate(location.href, "replace");
        else if (!$('moments').hidden) void loadMoments();
    };
    $("language").onchange = changeLanguage;
    $("language").oninput = changeLanguage;
    media.addEventListener("change", theme);
    queueMicrotask(() => navigate(location.href, "replace"));
    $("theme").onclick = () => {
        mode = modes[(modes.indexOf(mode) + 1) % modes.length];
        localStorage.setItem("dawn.site.theme", mode);
        theme();
    };
    async function api(path, method = "GET", data, full = false) {
        const response = await fetch(path, {
            method,
            signal: AbortSignal.timeout(path === "/api/moment-drafts" ? 120000 : 15000),
            credentials: "omit",
            headers: {
                "Content-Type": "application/json",
                "Accept-Language": language,
                ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
            ...(data === undefined ? {} : { body: JSON.stringify(data) }),
        });
        const result = response.status === 204 ? {} : await response.json();
        if (!response.ok)
            throw Object.assign(new Error(result.message || t("请求失败")), {
                status: response.status,
            });
        return full ? result : result.data;
    }
    function authMode() {
        $("auth-title").textContent = t(
            register ? (user?.type === "guest" ? "完善信息" : "注册") : "登录",
        );
        $("submit").textContent = t(register ? "注册" : "登录");
        $("toggle").textContent = t(
            register ? "已有账号，去登录" : "还没有账号，去注册",
        );
        $("nickname-label").hidden = !register;
        $("password").autocomplete = register
            ? "new-password"
            : "current-password";
    }
    function profile() {
        $("community-name").textContent = user?.nickname || t("欢迎来到黎明社区");
        $("community-id").textContent = user?.unique_id ? `#${user.unique_id}` : t("登录后参与交流");
        $("community-region").textContent = user?.region || "";
        $("community-avatar").replaceChildren(avatar(user || {}));
        $("account").textContent =
            user?.type === "guest"
                ? t("完善信息")
                : user?.nickname || user?.unique_id || t("登录");
        $("logout").hidden = !user || user.type === "guest";
    }
    let communityRegion = "";
    async function loadCommunityRegion() {
        const location = $("moment-location").value;
        try {
            const result = await api("/api/moment-location");
            user = user ? { ...user, region: result.region } : user;
            communityRegion = result.region === "未知" ? "" : result.region;
            $("community-region").textContent = communityRegion;
            if (!editingId && (!location || location === "未知") && $("moment-location").value === location && communityRegion) $("moment-location").value = communityRegion;
            return result.region;
        } catch {}
    }
    async function resolveIdentity({ visitor_ticket }) {
        token = localStorage.getItem("dawn.member") || "";
        const previousToken = token;
        const version = accountVersion;
        let result;
        try {
            result = await api("/api/support-identities", "POST", {
                visitor_ticket,
            });
        } catch (error) {
            if (error.status !== 401 || version !== accountVersion) throw error;
            token = "";
            localStorage.removeItem("dawn.member");
            result = await api("/api/support-identities", "POST", {
                visitor_ticket,
            });
        }
        if (version !== accountVersion) return null;
        if (result.token) {
            token = result.token;
            localStorage.setItem("dawn.member", token);
        }
        user = result.user;
        profile();
        void loadCommunityRegion();
        if (!$('moments').hidden) {
            renderForumNav();
            if (previousToken && token !== previousToken) void loadMoments();
        }
        return result.support;
    }
    $("account").onclick = () => {
        if (user?.type === "member") return;
        register = user?.type === "guest";
        $("nickname").value = user?.nickname || "";
        authMode();
        $("error").textContent = "";
        $("auth").showModal();
    };
    $("toggle").onclick = () => {
        register = !register;
        authMode();
        $("error").textContent = "";
    };
    $("close").onclick = () => $("auth").close();
    $("form").onsubmit = async (event) => {
        event.preventDefault();
        $("submit").disabled = true;
        $("error").textContent = "";
        try {
            const result = await api(
                register
                    ? "/api/account/registrations"
                    : "/api/account/sessions",
                "POST",
                {
                    username: $("username").value,
                    password: $("password").value,
                    nickname: $("nickname").value,
                },
            );
            ++accountVersion;
            token = result.token;
            user = result.user;
            localStorage.setItem("dawn.member", token);
            $("password").value = "";
            $("auth").close();
            profile();
            await window.DawnSupport.syncIdentity();
            if (!$('moments').hidden) void loadMoments();
        } catch (error) {
            $("error").textContent = error.message;
        } finally {
            $("submit").disabled = false;
        }
    };
    $("logout").onclick = async () => {
        ++accountVersion;
        try {
            await api("/api/account/session", "DELETE");
        } finally {
            token = "";
            user = undefined;
            localStorage.removeItem("dawn.member");
            profile();
            await window.DawnSupport?.reset();
            if (!$('moments').hidden) void loadMoments();
        }
    };
    addEventListener("storage", (event) => {
        if (event.key === "dawn.site.locale" && event.newValue && event.newValue !== language) {
            locale(event.newValue);
            if (!documentation.hidden) navigate(location.href, "replace");
            else if (!$("moments").hidden) void loadMoments();
        }
        if (event.key !== "dawn.member" || (event.newValue || "") === token)
            return;
        ++accountVersion;
        token = event.newValue || "";
        void window.DawnSupport?.syncIdentity().catch((error) => {
            $("error").textContent = error.message;
        });
    });
    $("support").onclick = (event) => {
        event.stopPropagation();
        void window.DawnSupport?.open();
    };
    const boards = { general: ["自由交流", "chat"], help: ["问题求助", "help"], showcase: ["作品分享", "grid"], ai: ["AI趣闻", "spark"] };
    let momentVersion = 0;
    function memberRequired() {
        if (user?.type === "member") return true;
        $("account").click();
        return false;
    }
    function momentButton(label, action, symbol) {
        const button = element("button", "moment-action");
        if (symbol) button.append(icon(symbol));
        button.append(element("span", "", label));
        button.type = "button";
        button.onclick = async () => {
            button.disabled = true;
            try { await action(button); } catch (error) { $("moment-error").textContent = error.message; }
            finally { button.disabled = false; }
        };
        return button;
    }
    async function deleteMoment(row) {
        if (!confirm(t("确认删除？"))) return;
        await api(`/api/moments/${row.id}`, "DELETE");
        if (location.pathname === `/moments/${row.id}`) navigate("/");
        else await loadMoments();
    }
    function avatar(row) {
        const node = element("span", "moment-avatar", (row.nickname || "D").slice(0, 1));
        if (row.avatar) {
            const image = element("img"); image.src = row.avatar; image.alt = ""; image.loading = "lazy"; node.replaceChildren(image);
        }
        return node;
    }
    function author(row) {
        const box = element("div", "moment-author"), person = row.is_anonymous ? element("span", "", t("匿名用户")) : row.user_type === "ai" ? element("span", "", row.nickname || "AI") : momentButton(row.nickname || row.unique_id, () => openPeer(row.unique_id));
        person.className = "author-name"; person.title = t("发消息");
        const info = element("div"), line = element("div", "author-line");
        line.append(person, element("span", "author-region", t(row.region || "未知")));
        if (row.user_type === "ai") line.append(element("span", "moment-badge", "AI"));
        info.append(line, element("time", "moment-meta", new Date(row.created_at).toLocaleString(language === "zh" ? "zh-CN" : "en-US")));
        box.append(avatar(row), info);
        return box;
    }
    function forumLink(changes) {
        const params = new URLSearchParams(location.search);
        params.delete("page");
        for (const [key, value] of Object.entries(changes)) value ? params.set(key, value) : params.delete(key);
        return `/${params.size ? `?${params}` : ""}`;
    }
    function renderForumNav() {
        const params = new URLSearchParams(location.search), board = params.get("board"), mine = params.get("author");
        const list = $("forum-boards"); list.replaceChildren();
        for (const [key, [label, symbol]] of [["", ["全部", "grid"]], ...Object.entries(boards)]) {
            const link = element("a", key === (board || "") ? "active" : "");
            link.href = forumLink({ board: key, author: "" }); link.append(icon(symbol), element("span", "", t(label))); list.append(link);
        }
        $("forum-query").value = params.get("query") || "";
        $("forum-title").textContent = mine ? t("我发的") : board && boards[board] ? t(boards[board][0]) : t("网友动态");
        const tabs = $("forum-tabs"); tabs.replaceChildren();
        for (const [key, label, symbol] of [["created_at", "新动态", "spark"], ["activity_at", "新评论", "chat"], ["hot", "热门贴", "fire"], ["mine", "我发的", "person"]]) {
            const selected = key === "mine" ? mine === user?.unique_id : !mine && key === (params.get("sort_by") || "created_at");
            const button = momentButton(t(label), () => {
                if (key === "mine" && !memberRequired()) return;
                navigate(forumLink({ sort_by: key === "mine" ? "created_at" : key, author: key === "mine" ? user.unique_id : "" }));
            }, symbol);
            button.classList.toggle("active", selected); button.setAttribute("role", "tab"); button.setAttribute("aria-selected", String(selected)); tabs.append(button);
        }
    }
    function momentImageURL(id, image) { return `/api/moments/${id}/images/${image}`; }
    let momentConfig = { emojis: {}, toolbar_hidden: [] };
    // 列表摘要：渲染后取纯文本，去掉 ** # [] 等格式符号与媒体，零 XSS 风险。
    function markdownToText(text) {
        const engine = window.DawnMomentMarkdown;
        if (!engine) return String(text || "").replace(/\s+/g, " ").trim();
        const node = element("div");
        node.innerHTML = engine.render(String(text || ""), {}, () => ({}));
        node.querySelectorAll("img,video,iframe,audio,table,[data-md-mask]").forEach(item => item.remove());
        return (node.textContent || "").replace(/\s+/g, " ").trim();
    }
    function renderSafeMarkdown(text, allowMedia, resolveAttachment) {
        const node = element("div", "markdown-body moment-content");
        const engine = window.DawnMomentMarkdown;
        node.innerHTML = engine.render(text, momentConfig.emojis, resolveAttachment);
        if (allowMedia) {
            engine.enhanceMarkdown(node);
            node.querySelectorAll("img").forEach(image => image.onclick = () => openImageViewer(image.src));
        } else node.querySelectorAll("img,video,iframe,audio,table,details,[data-md-mask]").forEach(item => item.remove());
        return node;
    }
    function openImageViewer(src) {
        const dialog = $("moment-image-viewer"), image = $("image-viewer-image");
        image.src = src; image.classList.remove("zoomed"); dialog.showModal();
    }
    $("image-viewer-close").onclick = () => $("moment-image-viewer").close();
    $("image-viewer-image").ondblclick = () => $("image-viewer-image").classList.toggle("zoomed");
    $("moment-image-viewer").onclick = (event) => { if (event.target === $("moment-image-viewer")) $("moment-image-viewer").close(); };
    async function loadOverview() {
        const data = await api("/api/moment-overview");
        for (const key of ["posts", "replies", "members"]) $(`stat-${key}`).textContent = data[key].toLocaleString();
        const list = $("forum-hot"); list.replaceChildren();
        for (const [index, row] of data.hot.entries()) {
            const link = element("a"); link.href = `/moments/${row.id}`;
            const text = element("div"); text.append(element("strong", "", row.title || markdownToText(row.content).slice(0, 60)), element("span", "moment-meta", `${row.comments} ${t("回复")} · ${row.nickname || row.unique_id}`));
            link.append(element("span", "hot-index", String(index + 1).padStart(2, "0")), text); list.append(link);
        }
        if (!data.hot.length) list.append(element("p", "widget-text", t("暂无热议")));
    }
    async function loadMoments() {
        if ($("moments").hidden) return;
        const version = ++momentVersion, params = new URLSearchParams(location.search);
        const id = location.pathname.match(/^\/moments\/(\d+)/)?.[1];
        const list = $("moment-list"), pages = $("moment-pages");
        renderForumNav(); $("forum-tabs").hidden = !!id;
        if (!list.children.length) list.replaceChildren(element("p", "moment-empty", t("加载中")));
        $("moment-error").textContent = "";
        void loadOverview().catch(() => {});
        try {
            const result = id ? { data: [await api(`/api/moments/${id}`)] } : await api(`/api/moments?${params}`, "GET", undefined, true);
            if (id) await markdownReady;
            if (version !== momentVersion) return;
            pages.replaceChildren();
            if (result.meta && result.meta.page > 1 && !result.data.length) {
                params.set("page", Math.max(1, Math.ceil(result.meta.total / result.meta.page_size)));
                navigate(`/?${params}`, "replace"); return;
            }
            list.replaceChildren(); list.classList.toggle("is-detail", !!id);
            if (id) {
                const back = element("a", "forum-back", `← ${t("返回动态")}`); back.href = "/"; list.append(back);
            }
            if (!result.data.length) {
                const empty = element("div", "forum-empty"); empty.append(icon("chat"), element("h2", "", t(params.size ? "没有找到相关主题" : "期待你的第一条分享")));
                if (params.size) { const clear = element("a", "widget-link", t("清除筛选")); clear.href = "/"; empty.append(clear); }
                list.append(empty);
            }
            for (const row of result.data) {
                const card = element("article", id ? "moment-detail" : "moment-row");
                const images = JSON.parse(row.images || "[]") || [], title = row.title || markdownToText(row.content).slice(0, 60);
                if (!id) {
                    const photo = element("a", "row-avatar"); photo.href = `/moments/${row.id}`; photo.append(avatar(row));
                    const body = element("div", "moment-row-body"), link = element("a", "moment-title", title); link.href = photo.href;
                    const meta = element("div", "moment-row-meta"), board = boards[row.board] || boards.general;
                    const boardLink = element("a", `board-tag board-${row.board || "general"}`, t(board[0])); boardLink.href = `/?board=${row.board || "general"}`;
                    meta.append(boardLink, element("span", "", row.is_anonymous ? t("匿名用户") : row.nickname || row.unique_id), element("span", "", t(row.region || "未知")), element("time", "", new Date(row.activity_at || row.created_at).toLocaleDateString(language === "zh" ? "zh-CN" : "en-US")));
                    if (row.user_type === "ai") meta.append(element("span", "moment-badge", "AI"));
                    const excerpt = element("p", "moment-excerpt", markdownToText(row.content).slice(0, 140));
                    if (!window.DawnMomentMarkdown) void markdownReady.then(() => {
                        if (excerpt.isConnected) excerpt.textContent = markdownToText(row.content).slice(0, 140);
                    }).catch(() => {});
                    body.append(link, excerpt, meta);
                    card.append(photo, body);
                    if (images.length) { const image = element("img", "row-thumbnail"); image.src = momentImageURL(row.id, images[0]); image.alt = ""; image.loading = "lazy"; image.onclick = () => openImageViewer(image.src); card.append(image); }
                    const actions = element("div", "row-actions");
                    for (const [kind, label, symbol] of [["flowers", "鲜花", "flower"], ["eggs", "鸡蛋", "egg"]]) actions.append(momentButton(`${t(label)} ${row[kind] || 0}`, async (button) => {
                        if (!memberRequired()) return;
                        await api(`/api/moments/${row.id}/reactions`, "POST", { kind }); row[kind]++; button.lastChild.textContent = `${t(label)} ${row[kind]}`;
                    }, symbol));
                    const replies = element("a", `reply-count${row.comments ? " has-replies" : ""}`); replies.href = photo.href; replies.title = `${t("回复")} ${row.comments}`; replies.append(icon("chat"), element("span", "", String(row.comments))); actions.append(replies);
                    if (row.can_manage) {
                        actions.append(momentButton(t("编辑"), () => openMomentEditor(row), "edit"));
                        actions.append(momentButton(t("删除"), () => deleteMoment(row), "trash"));
                    }
                    card.append(actions);
                } else {
                    document.title = `${title} · Dawn`;
                    const inline = new Set();
                    const content = renderSafeMarkdown(row.content, row.format === "post", (attachment) => {
                        if (!Object.hasOwn(row.image_names || {}, attachment)) return;
                        inline.add(Number(attachment));
                        return { url: momentImageURL(row.id, attachment), name: row.image_names[attachment] };
                    });
                    card.append(element("h1", "detail-title", title), author(row), content);
                    const galleryImages = images.filter(image => !inline.has(Number(image)));
                    if (galleryImages.length) {
                        const gallery = element("div", "moment-images");
                        for (const image of galleryImages) {
                            const link = element("a"); link.href = momentImageURL(row.id, image); link.onclick = (event) => { event.preventDefault(); openImageViewer(link.href); };
                            const img = element("img"); img.src = link.href; img.alt = ""; img.loading = "lazy"; link.append(img); gallery.append(link);
                        }
                        card.append(gallery);
                    }
                    if (row.location) { const place = element("p", "moment-location"); place.append(icon("pin"), document.createTextNode(row.location)); card.append(place); }
                    const actions = element("div", "moment-actions");
                    for (const [kind, label, symbol] of [["flowers", "鲜花", "flower"], ["eggs", "鸡蛋", "egg"]]) actions.append(momentButton(`${t(label)} ${row[kind] || 0}`, async (button) => {
                        if (!memberRequired()) return;
                        await api(`/api/moments/${row.id}/reactions`, "POST", { kind }); row[kind]++; button.lastChild.textContent = `${t(label)} ${row[kind]}`;
                    }, symbol));
                    actions.append(momentButton(t("复制链接"), async (button) => { await navigator.clipboard.writeText(location.href); button.lastChild.textContent = t("已复制"); }, "link"));
                    if (row.can_manage) actions.append(momentButton(t("编辑"), () => openMomentEditor(row), "edit"));
                    if (row.can_manage) actions.append(momentButton(t("删除"), () => deleteMoment(row), "trash"));
                    card.append(actions);
                }
                list.append(card);
                if (id) await loadComments(list, row);
            }
            if (result.meta && result.meta.total > result.meta.page_size) {
                const totalPages = Math.max(1, Math.ceil(result.meta.total / result.meta.page_size)), page = result.meta.page;
                if (page > 1) { const prev = element("a", "page-button", "←"); prev.href = forumLink({ page: String(page - 1) }); prev.title = t("上一页"); pages.append(prev); }
                for (const number of [...new Set([1, ...Array.from({ length: 5 }, (_, i) => page - 2 + i), totalPages])].filter(n => n > 0 && n <= totalPages).sort((a,b) => a-b)) {
                    const link = element("a", `page-button${number === page ? " active" : ""}`, number); link.href = forumLink({ page: String(number) }); if (number === page) link.setAttribute("aria-current", "page"); pages.append(link);
                }
                if (page < totalPages) { const next = element("a", "page-button", "→"); next.href = forumLink({ page: String(page + 1) }); next.title = t("下一页"); pages.append(next); }
            }
        } catch (error) { if (version === momentVersion) { list.replaceChildren(); $("moment-error").textContent = error.message; } }
    }
    let commentEditor = null;
    async function loadComments(container, post) {
        const version = momentVersion;
        const section = element("section", "moment-comments"), heading = element("h2", "comments-heading", `${t("评论")} ${post.comments}`);
        section.append(heading);
        const id = post.id, target = location.hash.match(/^#comment-(\d+)$/)?.[1];
        let page = 1, reply = 0;
        const entries = element("div"), more = element("div", "moment-pages");
        async function fetchPage() {
            const result = await api(`/api/moments/${id}/comments?page=${page}${target && page === 1 ? `&comment_id=${target}` : ""}`, "GET", undefined, true);
            if (version !== momentVersion) return;
            page = result.meta.page;
            for (const row of result.data) {
                const item = element("article", "moment-comment"); item.id = `comment-${row.id}`; item.append(author(row));
                const content = renderSafeMarkdown(row.content, true, (attachment) => {
                    if (!Object.hasOwn(row.image_names || {}, attachment)) return;
                    return {
                        url: `/api/moments/${id}/comments/${row.id}/images/${attachment}`,
                        name: row.image_names[attachment]
                    };
                });
                if (row.reply_id) content.prepend(element("span", "moment-reply", `${t("回复")} ${row.reply_name || "#" + row.reply_id}：`));
                item.append(content);
                const actions = element("div", "comment-actions");
                actions.append(momentButton(t("回复"), () => { if (!memberRequired()) return; reply = row.id; cancel.hidden = false; hint.textContent = `${t("回复")} ${row.nickname || row.unique_id}`; commentEditor.focus(); }, "chat"));
                if (user?.unique_id === row.unique_id) actions.append(momentButton(t("删除"), async () => { if (confirm(t("确认删除？"))) { await api(`/api/moments/${id}/comments/${row.id}`, "DELETE"); await loadMoments(); } }, "trash"));
                item.append(actions); entries.append(item);
            }
            more.replaceChildren();
            if (page * result.meta.page_size < result.meta.total) more.append(momentButton(t("下一页"), async () => { page++; await fetchPage(); }));
            if (target) {
                const found = document.getElementById(`comment-${target}`);
                if (found) { found.classList.add("highlighted"); found.scrollIntoView({ block: "center" }); }
                else $("moment-error").textContent = t("评论不存在或已删除");
            }
        }
        // 评论同样用完整编辑器：支持 md、图片、粘贴/拖入上传与我的历史图。
        const form = element("form", "moment-comment-form"), host = element("div", "comment-editor-shell");
        const send = element("button", "primary", t("发送")); send.type = "submit";
        const hint = element("span", "comment-reply-hint");
        const cancel = momentButton(t("取消回复"), () => { reply = 0; hint.textContent = ""; cancel.hidden = true; }); cancel.hidden = true;
        const tools = element("div", "comment-tools"); tools.append(hint, cancel, send);
        commentEditor?.destroy();
        commentEditor = window.DawnMomentEditor.create(host, {
            render: renderEditorMarkdown,
            enhance: enhanceEditorMarkdown,
            maxLength: 1024,
            pasteImages: true,
            imageLibrary: momentImageLibrary,
            imageUploader: momentImageUploader,
            previewImage: openImageViewer,
            config: () => ({ hidden: ["media"] }),
            locale: language === "en" ? "en" : "zh"
        });
        form.append(host, tools);
        form.onsubmit = async (event) => {
            event.preventDefault();
            if (!memberRequired()) return;
            const content = commentEditor.getValue();
            if (!content.trim()) return;
            send.disabled = true;
            try { await api(`/api/moments/${id}/comments`, "POST", { content, reply_id: reply }); commentEditor.setValue(""); reply = 0; hint.textContent = ""; cancel.hidden = true; await loadMoments(); }
            catch (error) { $("moment-error").textContent = error.message; }
            finally { send.disabled = false; }
        };
        section.append(entries, more, form); container.append(section); await fetchPage();
    }
    $("forum-search").onsubmit = (event) => { event.preventDefault(); navigate(forumLink({ query: $("forum-query").value.trim() })); };
    $("moment-ai").onclick = async () => {
        if (!memberRequired()) return;
        const button = $("moment-ai"); button.disabled = true; button.textContent = t("正在生成");
        try {
            const result = await api("/api/moment-drafts", "POST", { path: location.pathname });
            await openNewMoment(); momentEditor.setValue(result.content); $("moment-title").value = result.content.split("\n")[0].slice(0, 60); $("moment-board").value = "ai"; saveMomentDraft();
        } catch (error) { alert(error.message); }
        finally { button.disabled = false; button.textContent = t("AI写动态"); }
    };
    let chats = [], activeChat = 0, chatSocket, chatTimer, chatVersion = 0, chatResume = "", chatRetry = 0;
    let chatRows = [], renderedChat = 0, chatDrafts = {}, chatHasEarlier = false;
    const chatError = (error) => { $("chat-error").textContent = error.message; };
    async function loadChats() {
        const version = chatVersion;
        const rows = await api("/api/community-chats");
        if (version !== chatVersion) return;
        chats = rows;
        if (!chats.some((row) => row.id === activeChat)) activeChat = chats[0]?.id || 0;
        const list = $("chat-list"); list.replaceChildren();
        for (const row of chats) {
            const button = element("button", row.id === activeChat ? "active" : "");
            const title = element("span", "chat-peer", row.type === "room" ? t("聊天大厅") : row.nickname || row.peer);
            if (row.unread) title.append(element("span", "chat-unread", row.unread > 99 ? "99+" : row.unread));
            const preview = element("span", "chat-preview");
            if (window.DawnMomentMarkdown) {
                const plain = element("div");
                plain.innerHTML = window.DawnMomentMarkdown.render(row.content || "");
                plain.querySelectorAll("[data-md-mask]").forEach(mask => mask.textContent = "•••");
                preview.textContent = plain.textContent;
            } else preview.textContent = row.content;
            button.append(title, preview);
            button.type = "button";
            button.onclick = () => { activeChat = row.id; void loadChats().then(() => loadChatMessages()).catch(chatError); };
            list.append(button);
        }
        const unread = chats.reduce((sum, row) => sum + row.unread, 0);
        $("nav-chat-unread").hidden = !unread;
        $("nav-chat-unread").textContent = unread > 99 ? "99+" : unread;
        const current = chats.find((row) => row.id === activeChat);
        $("chat-title").textContent = current?.type === "room" ? t("聊天大厅") : current?.nickname || current?.peer || "";
        $("chat-input").value = chatDrafts[activeChat] || "";
        $("chat-input").disabled = user?.type !== "member";
        $("chat-input").placeholder = t(user?.type === "member" ? "输入消息" : "登录后参与聊天");
        $("chat-send").disabled = user?.type !== "member";
    }
    async function loadChatMessages(earlier = false) {
        if (!activeChat) return;
        const id = activeChat, version = chatVersion;
        const before = earlier && chatRows.length ? `&before_id=${chatRows[0].id}` : "";
        const rows = await api(`/api/community-chats/${id}/messages?page_size=50${before}`);
        if (id !== activeChat || version !== chatVersion) return;
        const list = $("chat-messages"), oldHeight = list.scrollHeight, oldTop = list.scrollTop;
        const atBottom = renderedChat !== id || oldHeight - oldTop - list.clientHeight < 70;
        if (earlier || renderedChat !== id) chatHasEarlier = rows.length === 50;
        if (renderedChat !== id) chatRows = [];
        renderedChat = id;
        chatRows = [...new Map([...chatRows, ...rows].map(row => [row.id, row])).values()].sort((a,b) => a.id-b.id);
        list.replaceChildren();
        if (chatHasEarlier) list.append(momentButton(t("更早的消息"), () => loadChatMessages(true)));
        if (!chatRows.length) list.append(element("p", "moment-empty", t("暂无消息，来打个招呼吧")));
        for (const row of chatRows) {
            const message = element("div", `chat-message${row.sender_subject === user?.unique_id ? " mine" : ""}`);
            const body = element("div", "markdown-body");
            const engine = window.DawnMomentMarkdown;
            if (engine) { body.innerHTML = engine.render(row.content || "", momentConfig.emojis); engine.enhanceMarkdown(body); }
            else body.textContent = row.content || "";
            message.append(element("strong", "", row.nickname || row.sender_subject), body);
            if (row.type === "event" && row.payload?.moment_id) {
                const link = element("button", "chat-card-link", t(row.payload.kind === "comment" ? "查看评论" : "查看动态"));
                link.type = "button";
                link.onclick = () => { $("community-chat").hidden = true; navigate(`/moments/${row.payload.moment_id}${row.payload.comment_id ? `#comment-${row.payload.comment_id}` : ""}`); };
                message.append(link);
            }
            list.append(message);
        }
        list.scrollTop = earlier ? oldTop + list.scrollHeight - oldHeight : atBottom ? list.scrollHeight : oldTop;
        if (!earlier && chatRows.length && !$("community-chat").hidden) {
            await api(`/api/community-chats/${id}/read`, "PUT", { message_id: chatRows.at(-1).id });
            await loadChats();
        }
    }
    async function openPeer(uniqueID) {
        if (!memberRequired() || uniqueID === user.unique_id) return;
        const chat = await api("/api/community-chats", "POST", { unique_id: uniqueID });
        activeChat = chat.id;
        await openCommunityChat();
    }
    async function openCommunityChat() {
        if (!user) { $("account").click(); return; }
        window.DawnSupport?.close(); $("community-chat").hidden = false; $("chat-error").textContent = "";
        try { await loadChats(); await loadChatMessages(); } catch (error) { chatError(error); }
    }
    $("nav-chat").onclick = openCommunityChat;
    document.querySelectorAll("[data-open-chat]").forEach(button => button.onclick = openCommunityChat);
    $("chat-close").onclick = () => { $("community-chat").hidden = true; };
    document.addEventListener("pointerdown", (event) => {
        const chat = $("community-chat");
        if (innerWidth > 600 && !chat.hidden && !chat.contains(event.target)) chat.hidden = true;
    });
    $("chat-input").oninput = () => { chatDrafts[activeChat] = $("chat-input").value; };
    $("chat-form").onsubmit = async (event) => {
        event.preventDefault(); if (!activeChat || !memberRequired()) return;
        const input = $("chat-input"), content = input.value, id = activeChat, button = $("chat-send");
        button.disabled = true; $("chat-error").textContent = "";
        try {
            await api(`/api/community-chats/${id}/messages`, "POST", { content, request_id: crypto.randomUUID() });
            chatDrafts[id] = "";
            if (activeChat === id) { input.value = ""; await loadChatMessages(); }
        } catch (error) { chatError(error); } finally { button.disabled = false; }
    };
    async function connectCommunity(version) {
        if (version !== chatVersion || !user) return;
        try {
            await loadChats();
            const ticket = await window.DawnSupport.createRealtimeTicket(chatResume);
            if (version !== chatVersion) return;
            const url = new URL(ticket.ws_url, location.href); url.protocol = url.protocol === "https:" ? "wss:" : url.protocol === "http:" ? "ws:" : url.protocol;
            const socket = new WebSocket(url); chatSocket = socket;
            socket.onopen = () => socket.send(JSON.stringify({ type: "auth", ticket: ticket.ticket }));
            socket.onmessage = (event) => {
                if (version !== chatVersion) return;
                const packet = JSON.parse(event.data);
                if (packet.type === "ready") { chatResume = packet.data.resume; chatRetry = 0; }
                if (packet.type === "resync_required" || packet.type.startsWith("chat.message.")) {
                    void loadChats().then(() => { if (!$("community-chat").hidden) return loadChatMessages(); }).catch(chatError);
                }
                if (packet.id) socket.send(JSON.stringify({ type: "ack", id: packet.id, channel: packet.channel }));
            };
            socket.onclose = () => { if (version === chatVersion) chatTimer = setTimeout(() => connectCommunity(version), Math.min(30000, 500 * 2 ** chatRetry++)); };
            socket.onerror = () => socket.close();
        } catch {
            chatResume = "";
            if (version === chatVersion) chatTimer = setTimeout(() => connectCommunity(version), Math.min(30000, 500 * 2 ** chatRetry++));
        }
    }
    function resetCommunity() {
        ++chatVersion; clearTimeout(chatTimer); chatSocket?.close(); chatResume = ""; chatRetry = 0; activeChat = 0; renderedChat = 0; chatRows = []; chatDrafts = {};
        $("community-chat").hidden = true; $("nav-chat-unread").hidden = true;
        if (user) void connectCommunity(chatVersion);
    }
    $("moment-compose").onclick = async () => {
        if (!memberRequired()) return;
        const button = $("moment-compose");
        button.disabled = true;
        try { await openNewMoment(); }
        catch (error) { $("moment-error").textContent = error.message; }
        finally { button.disabled = false; }
    };
    let momentFormat = "moment", editingId = 0, currentImages = [], openingMoment = 0;
    let loadingForm = false, galleryVersion = 0, galleryUploading = false, gallerySearchVersion = 0;
    let editingImageNames = {};
    const imageMetadata = new Map();
    const draftKey = "dawn.moment.drafts";
    const momentDrafts = (() => { try { return JSON.parse(localStorage.getItem(draftKey)) || {}; } catch { return {}; } })();
    function saveMomentDraft() {
        if (editingId || loadingForm) return;
        momentDrafts[momentFormat] = { title: $("moment-title").value, content: momentEditor.getValue(), board: $("moment-board").value, location: $("moment-location").value, is_anonymous: $("moment-anonymous").checked };
        localStorage.setItem(draftKey, JSON.stringify(momentDrafts));
    }
    function momentImageContentURL(id) { return `/api/moment-images/${id}/content`; }
    function editorConfig() {
        const hidden = [...(momentConfig.toolbar_hidden || [])];
        if (momentFormat === "moment") hidden.push("image", "media");
        return { hidden };
    }
    function renderEditorMarkdown(text, post = 0, names = {}) {
        return window.DawnMomentMarkdown.render(text, momentConfig.emojis, id =>
            Object.hasOwn(names, id)
                ? { url: momentImageURL(post, id), name: names[id] }
                : { url: momentImageContentURL(id) });
    }
    function enhanceEditorMarkdown(root) {
        window.DawnMomentMarkdown.enhanceMarkdown(root);
        root.querySelectorAll("img[data-moment-attachment]").forEach(img => {
            const id = img.dataset.momentAttachment;
            if (img.alt) return;
            if (!imageMetadata.has(id)) {
                imageMetadata.set(id, api(`/api/moment-images/${id}`).catch(() => {
                    imageMetadata.delete(id);
                    return null;
                }));
            }
            imageMetadata.get(id).then(row => { if (img.isConnected && row?.name) img.alt = row.name; });
        });
    }
    async function fingerprint(file) {
        const bytes = await file.arrayBuffer();
        const sha256 = [...new Uint8Array(await crypto.subtle.digest("SHA-256", bytes))].map(value => value.toString(16).padStart(2, "0")).join("");
        return { md5: SparkMD5.ArrayBuffer.hash(bytes), sha256 };
    }
    const momentImageUploader = {
        fingerprint,
        toMarkdown: image => `[attachment://${image.id}]`,
        check: async (files, prints) => {
            const hits = [];
            for (let i = 0; i < files.length; i++) {
                const result = await api("/api/moment-image-checks", "POST", { name: files[i].name, size: files[i].size, ...prints[i] });
                if (result.exists) imageMetadata.set(String(result.id), Promise.resolve(result));
                hits.push(result.exists ? { id: result.id, url: momentImageContentURL(result.id), name: result.name } : null);
            }
            return hits;
        },
        upload: async (files, prints) => {
            const rows = [];
            for (let i = 0; i < files.length; i++) {
                const body = new FormData(); body.append("file", files[i]); body.append("md5", prints[i].md5); body.append("sha256", prints[i].sha256);
                const response = await fetch("/api/moment-images", { method: "POST", headers: { Authorization: `Bearer ${token}`, "Accept-Language": language }, body });
                const result = await response.json();
                if (!response.ok) throw new Error(result.message || t("请求失败"));
                imageMetadata.set(String(result.data.id), Promise.resolve(result.data));
                rows.push({ id: result.data.id, url: momentImageContentURL(result.data.id), name: result.data.name });
            }
            return rows;
        }
    };
    const momentImageLibrary = {
        list: async (keyword) => {
            const rows = await api(`/api/moment-images/mine?keyword=${encodeURIComponent(keyword || "")}`);
            for (const row of rows || []) imageMetadata.set(String(row.id), Promise.resolve(row));
            return rows || [];
        }
    };
    function renderGallery() {
        $("moment-image-preview").replaceChildren(...currentImages.map(id => {
            const cell = element("div", "gallery-cell");
            const image = element("img"); image.src = editingId ? momentImageURL(editingId, id) : momentImageContentURL(id); image.alt = "";
            if (editingId) image.onerror = () => { image.onerror = null; image.src = momentImageContentURL(id); };
            const preview = element("button"); preview.type = "button"; preview.title = t("预览");
            preview.append(image); preview.onclick = () => openImageViewer(image.src);
            const remove = element("button", "gallery-remove", "×"); remove.type = "button";
            remove.title = t("移除图片"); remove.ariaLabel = remove.title;
            remove.onclick = () => { currentImages = currentImages.filter(item => item !== id); renderGallery(); };
            cell.append(preview, remove);
            return cell;
        }));
    }
    function addGalleryImages(ids) {
        const next = [...new Set([...currentImages, ...ids])];
        if (next.length > 9) throw new Error(t("最多选择9张图片"));
        currentImages = next;
        renderGallery();
    }
    async function loadGallery() {
        const version = ++gallerySearchVersion;
        const host = $("moment-gallery-library"), status = $("moment-gallery-status");
        host.replaceChildren(); status.textContent = t("加载中");
        try {
            const rows = await momentImageLibrary.list($("moment-gallery-search").value.trim());
            if (version !== gallerySearchVersion) return;
            status.textContent = rows.length ? "" : t("暂无图片");
            for (const row of rows) {
                const button = element("button"); button.type = "button"; button.title = row.name;
                const image = element("img"); image.src = row.url; image.alt = row.name;
                button.append(image);
                button.onclick = () => {
                    try { addGalleryImages([row.id]); $("moment-gallery").close(); }
                    catch (error) { status.textContent = error.message; }
                };
                host.append(button);
            }
        } catch (error) { if (version === gallerySearchVersion) status.textContent = error.message; }
    }
    $("moment-gallery-open").onclick = () => {
        $("moment-gallery-search").value = "";
        $("moment-gallery").showModal();
        void loadGallery();
    };
    $("moment-gallery-close").onclick = () => $("moment-gallery").close();
    $("moment-gallery-search").onkeydown = event => {
        if (event.key === "Enter") { event.preventDefault(); void loadGallery(); }
    };
    $("moment-gallery-search").oninput = () => void loadGallery();
    $("moment-gallery-upload").onclick = () => $("moment-gallery-files").click();
    $("moment-gallery-files").onchange = async event => {
        const files = Array.from(event.target.files); event.target.value = "";
        if (!files.length || galleryUploading) return;
        const status = $("moment-gallery-status"), version = galleryVersion;
        galleryUploading = true; $("moment-gallery-upload").disabled = true;
        try {
            if (files.length + currentImages.length > 9) throw new Error(t("最多选择9张图片"));
            status.textContent = t("正在上传图片");
            const prints = await Promise.all(files.map(fingerprint));
            const hits = await momentImageUploader.check(files, prints);
            const missing = files.filter((_, index) => !hits[index]);
            const uploaded = await momentImageUploader.upload(missing, prints.filter((_, index) => !hits[index]));
            if (version !== galleryVersion) return;
            addGalleryImages(hits.map(hit => (hit || uploaded.shift()).id));
            $("moment-gallery").close();
        } catch (error) { if (version === galleryVersion) status.textContent = error.message; }
        finally { galleryUploading = false; $("moment-gallery-upload").disabled = false; }
    };
    function setMomentFormat(format) {
        momentFormat = format;
        const post = format === "post";
        $("format-moment").classList.toggle("active", !post);
        $("format-post").classList.toggle("active", post);
        $("moment-title-field").hidden = !post;
        $("moment-anonymous-field").hidden = !post;
        momentEditor.setOptions({ maxLength: post ? 10000 : 1024, pasteImages: post, config: editorConfig });
    }
    function loadComposeFields(row) {
        loadingForm = true;
        $("moment-title").value = row.title || "";
        $("moment-board").value = row.board || "general";
        $("moment-location").value = editingId ? row.location || "" : (row.location && row.location !== "未知" ? row.location : communityRegion);
        $("moment-anonymous").checked = !!row.is_anonymous;
        momentEditor.setValue(row.content || "");
        loadingForm = false;
    }
    function showCompose() {
        for (const [node, label] of [[$("moment-dialog-title"), editingId ? "编辑动态" : "发布动态"], [$("moment-submit"), editingId ? "保存修改" : "发布"]]) {
            node.dataset.t = label;
            node.textContent = t(label);
        }
        $("moment-publish-note").hidden = !!editingId;
        $("moment-form-error").textContent = "";
        renderGallery();
        $("moment-dialog").showModal();
    }
    async function openNewMoment() {
        await editorReady;
        ++openingMoment;
        ++galleryVersion;
        editingId = 0; editingImageNames = {}; currentImages = [];
        setMomentFormat("moment");
        loadComposeFields(momentDrafts.moment || {});
        showCompose();
        void loadCommunityRegion();
    }
    async function openMomentEditor(row) {
        const version = ++openingMoment;
        await editorReady;
        row = await api(`/api/moments/${row.id}`);
        if (version !== openingMoment) return;
        ++galleryVersion;
        editingId = row.id;
        editingImageNames = row.image_names || {};
        currentImages = JSON.parse(row.images || "[]") || [];
        setMomentFormat(row.format || "post");
        loadComposeFields(row);
        showCompose();
    }
    for (const format of ["moment", "post"]) {
        $(`format-${format}`).onclick = () => {
            if (format === momentFormat) return;
            saveMomentDraft();
            setMomentFormat(format);
            if (!editingId) loadComposeFields(momentDrafts[format] || {});
        };
    }
    const editorReady = markdownReady.then(() => {
        momentEditor = window.DawnMomentEditor.create($("moment-editor-shell"), {
            render: text => renderEditorMarkdown(text, editingId, editingImageNames),
            enhance: enhanceEditorMarkdown,
            imageLibrary: momentImageLibrary,
            imageUploader: momentImageUploader,
            mediaMarkup: (url, type) => `MEDIA::${type}::${url}`,
            mentions: async (query) => {
                const rows = await api(`/api/mentions/search?q=${encodeURIComponent(query)}`);
                return (rows || []).map((row) => ({ label: `${row.nickname || row.unique_id} · ${row.unique_id}`, value: row.unique_id }));
            },
            config: editorConfig,
            previewImage: openImageViewer,
            onChange: saveMomentDraft,
            onSaveDraft: () => { saveMomentDraft(); if (!editingId) $("moment-form-error").textContent = t("草稿已保存"); },
            locale: language === "en" ? "en" : "zh",
            placeholder: t("分享你的想法"),
            maxLength: 1024
        });
        return momentEditor;
    });
    void editorReady.catch(() => {});
    for (const [value, [label]] of Object.entries(boards)) { const option = element("option", "", t(label)); option.value = value; option.dataset.t = label; $("moment-board").append(option); }
    $("moment-locate").onclick = async () => {
        try { const result = await api("/api/moment-location"); $("moment-location").value = t(result.region); }
        catch (error) { $("moment-form-error").textContent = error.message; }
    };
    $("moment-close").onclick = () => $("moment-dialog").close();
    $("moment-dialog").addEventListener("close", () => {
        ++galleryVersion;
        editingId = 0;
        editingImageNames = {};
        currentImages = [];
        loadingForm = true;
        momentEditor.setValue("");
        loadingForm = false;
    });
    $("moment-form").onsubmit = async (event) => {
        event.preventDefault();
        const button = $("moment-submit"), id = editingId;
        if (button.disabled || galleryUploading || $("moment-gallery").open) return;
        button.disabled = true;
        try {
            const payload = { format: momentFormat, content: momentEditor.getValue(), title: $("moment-title").value, board: $("moment-board").value, location: $("moment-location").value, images: currentImages, is_anonymous: $("moment-anonymous").checked };
            const row = id ? await api(`/api/moments/${id}`, "PUT", payload) : await api("/api/moments", "POST", payload);
            if (!id) { delete momentDrafts[momentFormat]; localStorage.setItem(draftKey, JSON.stringify(momentDrafts)); }
            $("moment-dialog").close();
            navigate(`/moments/${id || row.id}`);
        } catch (error) { $("moment-form-error").textContent = error.message; }
        finally { button.disabled = false; }
    };
    (async () => {
        try {
            site = await api("/api/site");
            momentConfig = (await api("/api/moment-config").catch(() => null)) || momentConfig;
            momentEditor?.setOptions({ config: editorConfig });
            void loadCommunityRegion();
            $("support").hidden = !site.support_key;
            if (site.support_key)
                window.DawnSupport.init({
                    server: location.origin,
                    app: site.support_key,
                    theme: mode,
                    locale: language,
                    resolveIdentity,
                    menu: [
                        { label: (locale) => locale === "zh" ? "客服" : "Support", icon: "/site-assets/support.svg", onClick: (support) => support.open() },
                        { label: (locale) => locale === "zh" ? "聊天" : "Chat", icon: "/site-assets/chat.svg", onClick: openCommunityChat },
                    ],
                    onSessionChange: resetCommunity,
                    linkRules: [
                        {
                            from: "https://dawnim.github.io/dawn-simple/",
                            to: "/docs/",
                        },
                    ],
                    onNavigate: (url) => navigate(url),
                    closeOnMobileNavigate: true,
                });
            if (!documentation.hidden)
                documentation.contentWindow.postMessage(
                    { source: "dawn-site", type: "state" },
                    location.origin,
                );
        } catch {
            $("support").hidden = true;
        }
    })();
    $("site-startup")?.remove();
})();
