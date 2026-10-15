const createForm = document.getElementById("create_post");
const categorySelect = createForm.elements.category;
const errorText = createForm.querySelector(".form_error");

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
    errorText.textContent =
      "Could not load categories. Try to reload the page.";
  }
}
loadCategories();

// "submit" is an event of the form, not of the button
createForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  errorText.textContent = "";

  const title = createForm.elements.title.value;
  const desc = createForm.elements.description.value;
  const categoryID = Number(categorySelect.value);

  try {
    const response = await fetch("http://localhost:8080/posts", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title,
        category_id: categoryID,
        init_comment_text: desc,
      }),
    });
    if (!response.ok) {
      const error = new Error(`Error HTTP: ${response.status}`);
      error.status = response.status;
      throw error;
    }
    const post = await response.json();
    // go to the page of the created post
    window.location.href = `postInside.html?id=${post.id}`;
  } catch (error) {
    console.log("create post failed", error);
    if (error instanceof TypeError) {
      // fetch itself failed: the server is not reachable
      errorText.textContent = "The server is not available. Try again later.";
    } else if (error.status === 401 || error.status === 403) {
      errorText.textContent = "Please log in to create a post.";
    } else if (error.status === 400) {
      errorText.textContent =
        "Check the title, the category and the text of the post.";
    } else {
      errorText.textContent = "Something went wrong. Try again later.";
    }
  }
});

const navList = document.querySelector(".nav-list");

// finds the "Create post" item and highlights it; returns false if it isn't rendered yet
function markCurrentNavItem() {
  const link = navList.querySelector('a[href="createPost.html"]');
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
