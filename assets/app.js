const themeToggle = document.getElementById("theme-toggle");
const body = document.body;

function setTheme(theme) {
    if (!themeToggle) {
        return;
    }

    if (theme === "dark") {
        body.classList.add("dark-mode");
        themeToggle.textContent = "Light";
        themeToggle.setAttribute("aria-pressed", "true");
    } else {
        body.classList.remove("dark-mode");
        themeToggle.textContent = "Dark";
        themeToggle.setAttribute("aria-pressed", "false");
    }
}

const savedTheme = localStorage.getItem("theme");
if (savedTheme) {
    setTheme(savedTheme);
} else if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
    setTheme("dark");
} else {
    setTheme("light");
}

if (themeToggle) {
    themeToggle.addEventListener("click", () => {
        if (body.classList.contains("dark-mode")) {
            setTheme("light");
            localStorage.setItem("theme", "light");
        } else {
            setTheme("dark");
            localStorage.setItem("theme", "dark");
        }
    });
}

window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", event => {
    if (!localStorage.getItem("theme")) {
        setTheme(event.matches ? "dark" : "light");
    }
});

const adSlots = document.querySelectorAll(".adsbygoogle");
if (adSlots.length > 0) {
    window.adsbygoogle = window.adsbygoogle || [];

    const adClient = adSlots[0].dataset.adClient;
    const script = document.createElement("script");
    script.async = true;
    script.crossOrigin = "anonymous";
    script.src = `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${adClient}`;
    script.addEventListener("load", () => {
        adSlots.forEach(() => window.adsbygoogle.push({}));
    });
    document.head.appendChild(script);
}
