import { getAvatarUrl } from "./avatar.js";
import { getUserID } from "./header.js";

const postsContainer = document.querySelector(".posts");
const sortSelect = document.querySelector('select[name="sort"]');
const likedSelect = document.querySelector('select[name="liked"]');
const categorySelect = document.querySelector('select[name="category"]');
const serverError = document.querySelector("#server-err")
const filterTags = document.querySelector("#filter-tags");
//for search
const searchForm = document.querySelector("#searchBoxContainer form");
const searchInput = document.querySelector("#searchString");
const seeMoreBtn = document.querySelector(".see-more");

let userID = null;
let activeTitle = "";
let currentPage = 1;

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

// Build the URL query parameters for the backend.
function buildQuery(page) {
  const params = new URLSearchParams({
    byLatest: sortSelect.value === "oldest" ? "false" : "true",
    page: String(page),
  });

  // Title search
  if (activeTitle) {
    params.set("title", activeTitle);
  }

  // Category filter
  if (categorySelect.value !== "") {
    params.set("categoryId", categorySelect.value);
  }

  // Reaction filter
  if (likedSelect.value !== "") {
    params.set("reaction", likedSelect.value);

    if (userID !== null) {
      params.set("userID", String(userID));
    }
  }

  return params.toString();
}

// Fetch posts, either replacing the existing list or appending another page.
async function getPostData({ append = false } = {}) {
  // append = true -> load the next page and add it below the existing posts
   serverError.hidden = true;
  const page = append ? currentPage + 1 : 1;

  try {
    const response = await fetch(
      `http://localhost:8080/posts?${buildQuery(page)}`
    );

    if (response.status === 404) {
      if (!append) {
        postsContainer.innerHTML = "<p>No posts found</p>";
      }

      seeMoreBtn.style.display = "none";
      return;
    }

    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status}`);
    }

    const postData = await response.json();

    if (!append) {
      postsContainer.innerHTML = "";
    }

    renderAllPosts(postData);

    currentPage = page;
    seeMoreBtn.style.display =
      page < postData.totalPages ? "block" : "none";
  } catch (error) {
    console.error("Could not load posts:", error);
    serverError.hidden = false;
    filterTags.hidden = true;
  }
}

// Search by title
searchForm.addEventListener("submit", (event) => {
  event.preventDefault();

  const value = searchInput.value.trim();

  if (value.length < 3 || value.length > 100) {
    searchInput.setCustomValidity(
      "Enter between 3 and 100 characters to search"
    );
    searchInput.reportValidity();
    return;
  }

  searchInput.setCustomValidity("");
  activeTitle = value;

  getPostData();
});

searchInput.addEventListener("input", () => {
  searchInput.setCustomValidity("");

  // Clearing the input also clears the active title filter.
  if (searchInput.value.trim() === "" && activeTitle !== "") {
    activeTitle = "";
    getPostData();
  }
});

// Sorting
sortSelect.addEventListener("change", () => {
  getPostData();
});

// Reaction
likedSelect.addEventListener("change", () => {
  getPostData();
});

// Category
categorySelect.addEventListener("change", () => {
  getPostData();
});

// Load category options from the backend.
async function loadCategories() {
  try {
    const response = await fetch("http://localhost:8080/categories");

    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status}`);
    }

    const categories = await response.json();

    for (const category of categories) {
      const option = document.createElement("option");
      option.value = category.id;
      option.textContent = category.category_name;
      categorySelect.append(option);
    }
  } catch (error) {
    console.error("Could not load categories:", error);
  }
}

// Render the posts.
function renderAllPosts(data) {
  for (const post of data.posts) {
    const postID = post.id;
    const userName = post.author.username;
    const date = formatDate(post.created_at);
    const title = post.title;
    const category = post.category_name;
    const likes = post.likes;
    const amountComments = post.commentIDs.length - 1;

    const newEl = `
      <a href="postInside.html?id=${postID}"
         aria-label="Link to the post"
         class="post">

        <div class="author">
          <div class="image-container">
            <img
              class="author_img"
              src="${getAvatarUrl(post.author)}"
              alt="Photo of ${userName}"
            >
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
      </a>
    `;

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

// Load the logged-in user's ID before loading posts.
// This is needed for the Liked and Disliked filters.
async function initialize() {
  try {
    userID = await getUserID();
  } catch (error) {
    console.error("Could not get the logged-in user:", error);
  }

  // Liked/Disliked filters need a logged-in user ID.
  if (userID == null) {
    for (const option of likedSelect.options) {
      if (option.value === "liked" || option.value === "disliked") {
        option.disabled = true;
      }
    }
  }

  await loadCategories();
  await getPostData();
}

seeMoreBtn.addEventListener("click", () => {
  getPostData({ append: true });
});

initialize();
