import { getUserID } from "./header.js";

const postsContainer = document.querySelector(".posts-of-user");

async function getMyPosts() {
  try {
    const params = new URLSearchParams(window.location.search);
    const userID = params.get("id") ?? (await getUserID());
    if (userID === null) {
      postsContainer.innerHTML = "<p>Log in to see your posts</p>";
      return;
    }
    const response = await fetch(
      `http://localhost:8080/posts?search=true&field=authorId&value=${userID}`,
    );
    // the backend answers 404 when the user has no posts at all
    if (response.status === 404) {
      postsContainer.innerHTML = "<p>No posts yet</p>";
      return;
    }
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    const postData = await response.json();
    renderPosts(postData);
  } catch (error) {
    console.log("the error was catched", error);
  }
}
getMyPosts();

function renderPosts(data) {
  for (let i = 0; i < data.posts.length; i++) {
    const postID = data.posts[i].id;
    const userName = data.posts[i].author.username;
    const date = formatDate(data.posts[i].created_at);
    const title = data.posts[i].title;
    const category = data.posts[i].category_name;
    const likes = data.posts[i].likes;
    const amountComments = data.posts[i].commentIDs.length - 1;
    const newEl = `
                    <a href="postInside.html?id=${postID}" aria-label="link to the post" class="post">
                    <div class="author">
                        <div class="image-container">
                            <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${userName}">
                        </div>
                        <div class="author-info">
                            <p class="username">${userName}</p>
                            <p class="date">${date}</p>
                        </div>
                    </div>
                    <h3 class="post_title">${title}</h3>
                    <div class="post-footer">
                        <div class="tags">
                            <div class="tags_item">${category}</div>
                        </div>
                        <div class="stats">
                            <p class="post_comments">${amountComments}</p>
                            <p class="post_likes">${likes}</p>
                        </div>
                    </div>
                </a>`;
    postsContainer.insertAdjacentHTML("beforeend", newEl);
  }
}

function formatDate(isoString) {
  const date = new Date(isoString);
  return date.toLocaleDateString(undefined, {
    day: "numeric",
    month: "long",
    year: "numeric",
  });
}

const navList = document.querySelector(".nav-list");

// finds the "My posts" item and highlights it; returns false if it isn't rendered yet
function markCurrentNavItem() {
  const link = navList.querySelector('a[href^="myPosts.html"]');
  if (!link) return false;
  link.closest("li").classList.add("choosed");
  return true;
}

// header.js renders the item later (after the /me request), so we wait for it
if (!markCurrentNavItem()) {
  const observer = new MutationObserver(() => {
    if (markCurrentNavItem()) observer.disconnect();
  });
  observer.observe(navList, { childList: true });
}
