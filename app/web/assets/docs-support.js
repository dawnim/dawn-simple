(() => {
    "use strict";
    if (parent === window) return;
    function notify(action = "replace") {
        const url = new URL(location.href);
        url.searchParams.delete("__content");
        replace(history.state, "", url);
        parent.postMessage(
            { source: "dawn-docs", type: "navigate", url: url.href, action },
            location.origin,
        );
    }
    function state() {
        parent.postMessage(
            {
                source: "dawn-docs",
                type: "state",
                title: document.title,
                locale: document.documentElement.lang.startsWith("zh") ? "zh" : "en",
                theme: document.documentElement.classList.contains("dark")
                    ? "dark"
                    : "light",
            },
            location.origin,
        );
    }
    // 外层维护浏览器历史，文档内继续使用VitePress导航，避免产生两份历史记录。
    const replace = history.replaceState.bind(history);
    for (const method of ["pushState", "replaceState"]) {
        history[method] = (data, title, url) => {
            replace(data, title, url);
            notify(method === "pushState" ? "push" : "replace");
        };
    }
    addEventListener("hashchange", () => notify());
    document.addEventListener(
        "click",
        (event) => {
            parent.postMessage(
                { source: "dawn-docs", type: "click" },
                location.origin,
            );
            const link = event.target.closest("a[href]");
            if (
                !link ||
                event.button ||
                event.metaKey ||
                event.ctrlKey ||
                event.shiftKey ||
                event.altKey ||
                link.target === "_blank"
            )
                return;
            const url = new URL(link.href);
            if (url.origin === location.origin && url.pathname === "/") {
                event.preventDefault();
                event.stopPropagation();
                parent.postMessage(
                    {
                        source: "dawn-docs",
                        type: "navigate",
                        url: url.href,
                        action: "push",
                    },
                    location.origin,
                );
            }
        },
        true,
    );
    new MutationObserver(state).observe(document.documentElement, {
        attributes: true,
        attributeFilter: ["class", "lang"],
    });
    new MutationObserver(state).observe(document.head, {
        childList: true,
        subtree: true,
        characterData: true,
    });
    addEventListener("message", (event) => {
        if (
            event.origin === location.origin &&
            event.source === parent &&
            event.data?.source === "dawn-site" &&
            event.data.type === "state"
        )
            state();
    });
    notify();
    state();
})();
