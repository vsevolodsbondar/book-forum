import { getUserID } from "./header.js";

const postsContainer = document.querySelector(".posts");
const categorySelect = document.querySelector("select[name='category']");
const searchStringValue = document.querySelector("select[name='searchString']")
//for search
const searchForm = document.querySelector("#searchBoxContainer form");
const searchInput = document.querySelector("#searchString");

const seeMoreBtn = document.querySelector(".see-more");
let currentPage = 1;
//get info about posts

// builds the query string from the current state of the controls
let byLatest = "true";
let reactionField = null; // "likedBy", "dislikedBy" or null
let userID = null;
function buildQuery(page) {
  const params = new URLSearchParams({ byLatest, page });

  if (reactionField) {
    params.set("search", "true");
    params.set("field", reactionField);
    params.set("value", userID);
  } else if (categorySelect.value !== "") {
    params.set("search", "true");
    params.set("field", "categoryId");
    params.set("value", categorySelect.value);
  } else if (searchInput.value.trim() !== "") {
    params.set("search", "true");
    params.set("field", "title");
    params.set("value", searchInput.value.trim());
  }

  return params.toString();
}

searchForm.addEventListener("submit", (event) => {
  event.preventDefault();

  const searchValue = searchInput.value.trim();

  if (searchValue.length < 3) {
    searchInput.setCustomValidity(
      "Enter at least 3 characters to search",
    );
    searchInput.reportValidity();
    return;
  }

  searchInput.setCustomValidity("");

  // Search by title replaces category/reaction filtering.
  categorySelect.value = "";
  reactionField = null;

  filters
    .querySelectorAll("button[data-filter]")
    .forEach((button) => button.classList.remove("choosed"));

  filters
    .querySelector(
      `[data-filter="${byLatest === "true" ? "new" : "old"}"]`,
    )
    .classList.add("choosed");

  getPostData();
});

searchInput.addEventListener("input", () => {
  searchInput.setCustomValidity("");
});

async function getPostData({ append = false } = {}) {
  // append = true -> load the next page and add it below the existing posts
  const page = append ? currentPage + 1 : 1;
  try {
    const response = await fetch(
      `http://localhost:8080/posts?${buildQuery(page)}`,
    );
    // the backend answers 404 when there are no posts at all
    if (response.status === 404) {
      postsContainer.innerHTML = "<p>No posts yet</p>";
      seeMoreBtn.style.display = "none";
      return;
    }
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    const postData = await response.json();
    if (!append) postsContainer.innerHTML = "";
    renderAllPosts(postData);
    currentPage = page;
    // show the button only while there are more pages
    seeMoreBtn.style.display = page < postData.totalPages ? "block" : "none";
  } catch (error) {
    console.log("the error was catched", error);
  }
}
getPostData();
seeMoreBtn.addEventListener("click", () => getPostData({ append: true }));

// filters above the list: New / Old
const filters = document.querySelector(".main-section > .tags");
filters.addEventListener("click", (event) => {
  const btn = event.target.closest("button[data-filter]");
  if (!btn) return;

  filters
    .querySelectorAll("button")
    .forEach((b) => b.classList.remove("choosed"));

  btn.classList.add("choosed");

  const filter = btn.dataset.filter;

  // Resetштп  other filters when switching the post filter.
  searchInput.value = "";
  categorySelect.value = "";

  reactionField =
    filter === "liked" ? "likedBy" : filter === "disliked" ? "dislikedBy" : null;
  if (reactionField) {
    categorySelect.value = ""; // category and liked/disliked can't be combined
  } else {
    byLatest = filter === "new" ? "true" : "false";
  }
  getPostData();
});

// "Liked posts" is shown only for logged in users
async function showLikedButton() {
  userID = await getUserID();
  if (userID !== null) {
    filters
      .querySelectorAll('[data-filter="liked"], [data-filter="disliked"]')
      .forEach((b) => (b.style.display = ""));
  }
}
showLikedButton();

// fill the category <select> with categories from the backend
async function loadCategories() {
  try {
    const response = await fetch("http://localhost:8080/categories");
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    const categories = await response.json();
    for (const category of categories) {
      const option = document.createElement("option");
      option.value = category.id;
      option.textContent = category.category_name;
      categorySelect.append(option);
    }
  } catch (error) {
    console.log("the error was catched", error);
  }
}
loadCategories();
// choosing a category asks the backend for posts of that category only
categorySelect.addEventListener("change", () => {
  searchInput.value = "";
  reactionField = null;

  filters
    .querySelectorAll("button")
    .forEach((b) => b.classList.remove("choosed"));
  filters
    .querySelector(`[data-filter="${byLatest === "true" ? "new" : "old"}"]`)
    .classList.add("choosed");
  getPostData();
});
//function for rendering all posts
function renderAllPosts(data) {
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
//formating of the date
function formatDate(isoString) {
  const date = new Date(isoString);
  return date.toLocaleDateString(undefined, {
    day: "numeric",
    month: "long",
    year: "numeric",
  });
}
