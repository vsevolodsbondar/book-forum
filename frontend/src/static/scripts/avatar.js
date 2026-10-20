export function getAvatarUrl(user) {
    if (!user?.profilepic) {
        return "src/static/imgs/default.jpg";
    }

    return `http://localhost:8080/avatars/${encodeURIComponent(user.profilepic)}`;
}

document.addEventListener("error", (event) => {
    const img = event.target;

    if (!(img instanceof HTMLImageElement)) return;
    if (!img.classList.contains("author_img")) return;
    if (img.dataset.fallbackAttempted === "true") return;

    console.log("Avatar failed to load:", img.src);

    img.dataset.fallbackAttempted = "true";
    img.src = "src/static/imgs/default.jpg";
}, true);
