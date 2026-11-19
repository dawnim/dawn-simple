import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { runInNewContext } from "node:vm";

test("文档内点击通过同源消息通知官网外壳", () => {
    const script = readFileSync(
        new URL("./assets/docs-support.js", import.meta.url),
        "utf8",
    );
    const messages = [],
        listeners = {};
    runInNewContext(script, {
        URL,
        location: new URL("https://site.test/docs/start.html"),
        history: { state: null, replaceState() {} },
        document: {
            documentElement: {
                lang: "zh-CN",
                classList: { contains: () => false },
            },
            head: {},
            addEventListener: (name, listener) => (listeners[name] = listener),
        },
        parent: {
            postMessage: (message, origin) => messages.push({ message, origin }),
        },
        window: {},
        MutationObserver: class {
            observe() {}
        },
        addEventListener: (name, listener) => (listeners[name] = listener),
    });
    messages.length = 0;
    listeners.click({ target: { closest: () => null } });
    assert.equal(messages[0].message.source, "dawn-docs");
    assert.equal(messages[0].message.type, "click");
    assert.equal(messages[0].origin, "https://site.test");
});
