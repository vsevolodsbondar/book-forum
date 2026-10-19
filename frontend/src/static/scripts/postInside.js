import { getUserID } from "./header.js";

const form = document.querySelector(".message");
//api to post by id
const params = new URLSearchParams(window.location.search);
const postID = params.get("id");
async function getPostData() {
  try {
    const userID = await getUserID();
    const response = await fetch(
      `http://localhost:8080/posts/${postID}/comments`,
      { credentials: "include" },
    );
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    const postData = await response.json();
    let isLoggedIn = true;
    let isOwner = false;
    if (userID === postData.comments[0].user_id) {
      isOwner = true;
    }
    if (userID === null) {
      isLoggedIn = false;
    }
    renderPostData(postData, isOwner, isLoggedIn, userID);
  } catch (error) {
    console.log("the error was catched", error);
  }
}
getPostData();

//render post and comments
function renderPostData(data, isOwner, isLoggedIn, userID) {
  const postDiscription = document.querySelector(".comment.init");
  postDiscription.id = data.comments[0].id;
  postDiscription.innerHTML = `
    <div class="author">
                    <div class="image-container">
                        <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${data.comments[0].user_for_comment.username}">
                    </div>
                    <div class="author-info">
                        <p class="username">${data.comments[0].user_for_comment.username}</p>
                        <p class="date">${formatDate(data.comments[0].updated_at)}</p>
                    </div>
                </div>
                <h3 class="post_title">${data.title}</h3>
                <p class="description">${data.comments[0].text}</p>
                <div class="post-footer">
                    <div class="tags">
                        <div class="tags_item">${data.category}</div>
                    </div>
                    ${isOwner ? '<button class="delete delete-post">Delete the post</button>' : ""}
                    <div class="stats">
                        <button class="likes ${data.comments[0].is_liked ? "choosed" : ""}">${data.comments[0].likes}</button>
                        <button class="dislikes ${data.comments[0].is_disliked ? "choosed" : ""}">${data.comments[0].dislikes}</button>
                    </div>
                </div>`;
  renderAllComments(data, userID, isLoggedIn);
  toDisableForm(isLoggedIn);
  if (isLoggedIn) {
    form.querySelector(".main_btn").addEventListener("click", (event) => {
      event.preventDefault();
      createComment(form, postID);
    });
  }
}
function renderComment(comment, isOwner, isLoggedIn) {
  // list of replies of the parent comment; null for top-level comments
  // and for comments to the post itself (the init article has no such list)
  const parentList = comment.parent_comment_id
    ? document.querySelector(`[id="${comment.parent_comment_id}"] > .replies-list`)
    : null;
  const newEl = `
        <article id="${comment.id}" class="comment${parentList ? " replies" : ""}">
                <div class="author">
                    <div class="image-container">
                        <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${comment.user_for_comment.username}">
                    </div>
                    <div class="author-info">
                        <p class="username">${comment.user_for_comment.username}</p>
                        <p class="date">${formatDate(comment.updated_at)}</p>
                    </div>
                </div>
                <p class="description">${comment.text}</p>
                <div class="post-footer">
                    <div class="stats">
                        <button class="likes ${comment.is_liked ? "choosed" : ""}">${comment.likes}</button>
                        <button class="dislikes ${comment.is_disliked ? "choosed" : ""}">${comment.dislikes}</button>
                    </div>
                    <div class="reply">
                        ${isLoggedIn && isOwner ? '<button class="blue_empty_btn btn_update">Edit</button>' : ""}
						            ${isLoggedIn && isOwner ? '<button class="delete btn_delete">Delete</button>' : ""}
                        ${isLoggedIn && !parentList ? '<button class="blue_empty_btn btn_reply">Reply</button>' : ""}
                    </div>
                </div>
                <div class="replies-list"></div>
            </article>`;
  // a reply goes inside the .replies-list of its parent, other comments go to the main list
  (parentList ?? document.querySelector(".comment-container")).insertAdjacentHTML(
    "beforeend",
    newEl,
  );
}
function renderAllComments(data, userID, isLoggedIn) {
  for (let i = 1; i < data.comments.length; i++) {
    const isCommentOwner = data.comments[i].user_id === userID;
    renderComment(data.comments[i], isCommentOwner, isLoggedIn);
  }
}
//to disable the form for comments if the user isn't logged in
function toDisableForm(isLoggedIn) {
  if (!isLoggedIn) {
    form.querySelector(".message_field").disabled = true;
    form.querySelector(".shadow_btn").disabled = true;
    form.querySelector(".main_btn").disabled = true;
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
//click function for creating a new comment
async function createComment(form, postID, parentID = null) {
  const textValue = form.querySelector(".message_field").value.trim();
  // don't send empty comment
  if (textValue === "") {
    return;
  }
  const data = {
    text: textValue,
    parent_comment_id: parentID,
  };
  try {
    const response = await fetch(
      `http://localhost:8080/posts/${postID}/comments`,
      {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json;charset=utf-8" },
        body: JSON.stringify(data),
      },
    );
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    const result = await response.json();
    console.log("Successful send:", result);
    // POST returns only the raw comment, so ask the server for the full one
    const fullResponse = await fetch(
      `http://localhost:8080/comments/${result.id}`,
      { credentials: "include" },
    );
    if (!fullResponse.ok) {
      throw new Error(`Error HTTP: ${fullResponse.status}`);
    }
    const fullComment = await fullResponse.json();

    form.querySelector(".message_field").value = "";
    renderComment(fullComment, true, true);
  } catch (error) {
    console.error("Error creating comment:", error);
  }
}
// click events on update button
const postInside = document.querySelector(".post-inside");
postInside.addEventListener("click", (event) => {
  const btn = event.target.closest(".btn_update");
  if (!btn) return;
  const comment = btn.closest(".comment");
  const footer = comment.querySelector(".post-footer");
  const description = comment.querySelector(".description");
  const elUpdate = `<div class="message update-form">
                <textarea class="message_field input" placeholder="">${description.textContent}</textarea>
                <div class="btn-container">
                    <button class="shadow_btn cancel">Cancel</button>
                    <button class="main_btn submit">Save</button>
                </div>
            </div>`;
  footer.style.display = "none";
  description.style.display = "none";
  description.insertAdjacentHTML("afterend", elUpdate);
  const cancelBtn = comment.querySelector(".update-form .cancel");
  const submitBtn = comment.querySelector(".update-form .submit");
  const updateForm = comment.querySelector(".update-form");
  cancelBtn.addEventListener("click", () => {
    footer.style.display = "flex";
    description.style.display = "block";
    updateForm.remove();
  });
  submitBtn.addEventListener("click", async () => {
    const trimmed = comment
      .querySelector(".update-form .message_field")
      .value.trim();
    if (trimmed === "") return;

    const closeForm = () => {
      footer.style.display = "flex";
      description.style.display = "block";
      updateForm.remove();
    };

    // nothing changed: backend would answer 400, so don't even ask
    if (trimmed === description.textContent) {
      closeForm();
      return;
    }

    try {
      const response = await fetch(
        `http://localhost:8080/comments/${comment.id}`,
        {
          method: "PATCH",
          credentials: "include",
          headers: { "Content-Type": "application/json;charset=utf-8" },
          body: JSON.stringify({ text: trimmed }),
        },
      );
      if (!response.ok) {
        throw new Error(`Error HTTP: ${response.status}`);
      }
      const result = await response.json();
      description.textContent = result.text;
      closeForm();
    } catch (error) {
      console.error("Error updating comment:", error);
    }
  });
});
// click on Reply: show a form under this comment
postInside.addEventListener("click", (event) => {
  const btn = event.target.closest(".btn_reply");
  if (!btn) return;

  const comment = btn.closest(".comment");

  // one reply form per comment: if it is already open, do nothing
  if (comment.querySelector(".reply-form")) return;

  const elReply = `<div class="message reply-form">
                <textarea class="message_field input" placeholder="Write your reply"></textarea>
                <div class="btn-container">
                    <button class="shadow_btn cancel">Cancel</button>
                    <button class="main_btn submit">Send</button>
                </div>
            </div>`;

  comment.insertAdjacentHTML("beforeend", elReply);

  const replyForm = comment.querySelector(".reply-form");
  replyForm.querySelector(".cancel").addEventListener("click", () => {
    replyForm.remove();
  });
  replyForm.querySelector(".submit").addEventListener("click", async () => {
    await createComment(replyForm, postID, Number(comment.id));
    replyForm.remove();
  });
});
// click on delete button of a comment
postInside.addEventListener("click", async (event) => {
  const btn = event.target.closest(".btn_delete");
  if (!btn) return;
  const comment = btn.closest(".comment");

  try {
    const response = await fetch(
      `http://localhost:8080/comments/${comment.id}`,
      { method: "DELETE", credentials: "include" },
    );
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }

    // the backend either deleted the comment or turned it into "Deleted message"
    const check = await fetch(`http://localhost:8080/comments/${comment.id}`, {
      credentials: "include",
    });
    if (check.status === 404) {
      comment.remove();
      return;
    }
    const left = await check.json();
    comment.querySelector(".description").textContent = left.text;
    comment.querySelector(".username").textContent =
      left.user_for_comment.username;
    comment
      .querySelectorAll(".btn_update, .btn_delete")
      .forEach((el) => el.remove());
  } catch (error) {
    console.error("Error deleting comment:", error);
  }
});

// click on delete button of the post
postInside.addEventListener("click", async (event) => {
  const btn = event.target.closest(".delete-post");
  if (!btn) return;

  try {
    const response = await fetch(`http://localhost:8080/posts/${postID}`, {
      method: "DELETE",
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    // the post (with its comments and likes) is gone, so there is nothing to show here
    window.location.href = "index.html";
  } catch (error) {
    console.error("Error deleting post:", error);
  }
});

//endpoint for likes/dislikes
const API = "http://localhost:8080";
function changeAmount(button, shift) {
  button.textContent = Number(button.textContent) + shift;
}
async function sendReaction(method, commentID, value) {
  const response = await fetch(`${API}/reaction`, {
    method,
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ comment_id: commentID, like: value }),
  });
  if (!response.ok) {
    const error = new Error("reaction failed");
    error.status = response.status;
    throw error;
  }
}
document.addEventListener("click", async (event) => {
  const button = event.target.closest(".likes, .dislikes");
  if (!button) return;
  const comment = button.closest(".comment");
  if (!comment) return;

  const likeBtn = comment.querySelector(".likes");
  const dislikeBtn = comment.querySelector(".dislikes");

  const isLike = button.classList.contains("likes");
  const value = isLike ? 1 : -1;
  const opposite = isLike ? dislikeBtn : likeBtn;
  const commentId = Number(comment.id);

  const wasChosen = button.classList.contains("choosed");
  const oppositeChosen = opposite.classList.contains("choosed");
  button.disabled = true;
  try {
    if (wasChosen) {
      // second click on the same button, so delete the reaction
      await sendReaction("DELETE", commentId, value);
      button.classList.remove("choosed");
      changeAmount(button, -1);
    } else if (oppositeChosen) {
      // click on the opposite button, so update the reaction
      await sendReaction("PATCH", commentId, value);
      opposite.classList.remove("choosed");
      changeAmount(opposite, -1);
      button.classList.add("choosed");
      changeAmount(button, 1);
    } else {
      // first click, so create the reaction
      await sendReaction("POST", commentId, value);
      button.classList.add("choosed");
      changeAmount(button, 1);
    }
  } catch (error) {
    if (error.status === 401 || error.status === 403) {
      window.location.href = "login.html";
    } else {
      console.error(error);
    }
  } finally {
    button.disabled = false;
  }
});
