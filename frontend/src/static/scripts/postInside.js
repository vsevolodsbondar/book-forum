import { getUserID } from "./header.js";
const form = document.querySelector(".message");
//api to post by id
const params = new URLSearchParams(window.location.search);
const postID = params.get("id");
async function getPostData(){
    try{
        const userID = await getUserID();
        const response = await fetch(`http://localhost:8080/posts/${postID}/comments`);
        if (!response.ok){
            throw new Error(`Error HTTP: ${response.status}`); 
        };
        const postData = await response.json();
        let isLoggedIn = true;
        let isOwner = false;
        if (userID == postData.comments[0].user_id){
            isOwner = true;
        };
        if (userID===null){
            isLoggedIn = false;
        }
        renderPostData(postData, isOwner, isLoggedIn);
        console.log(postData);
    } catch(error){
        console.log("the error was catched", error);
    }
};
getPostData();

//render post and comments
function renderPostData(data, isOwner, isLoggedIn){
    const postDiscription = document.querySelector(".comment.init")
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
                    ${isOwner ? "<button id='delete' class='blue_empty_btn'>Delete the post</button>": ""}
                    <div class="stats">
                        <button class="likes ${data.comments[0].is_liked ? "choosed" : ""}">${data.comments[0].likes}</button>
                        <button class="dislikes ${data.comments[0].is_disliked ? "choosed" : ""}">${data.comments[0].dislikes}</button>
                    </div>
                    
                </div>`;
                renderAllComments(data, isOwner, isLoggedIn);
                toDisableForm(isLoggedIn);
                if (isLoggedIn){
                    form.querySelector(".main_btn").addEventListener("click", (event)=>{
                        event.preventDefault();
                        createComment(form, postID);
                    });
                }
};
function renderAllComments(data, isOwner, isLoggedIn){
    const container = document.querySelector(".comment-container");
    for (let i=1; i<data.comments.length; i++){
        const name = data.comments[i].user_for_comment.username;
        const date = formatDate(data.comments[i].updated_at);
        const text = data.comments[i].text;
        const likes = data.comments[i].likes;
        const dislikes = data.comments[i].dislikes;
        const isLiked = data.comments[i].is_liked;
        const isDisliked = data.comments[i].is_disliked;
        const id = data.comments[i].id;
        const newEl = `
        <article id="${id}" class="comment">
                <div class="author">
                    <div class="image-container">
                        <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${name}">
                    </div>
                    <div class="author-info">
                        <p class="username">${name}</p>
                        <p class="date">${date}</p>
                    </div>
                </div>
                <p class="description">${text}</p>
                <div class="post-footer">
                    <div class="stats">
                        <button class="likes ${isLiked ? "choosed" : ""}">${likes}</button>
                        <button class="dislikes ${isDisliked ? "choosed" : ""}">${dislikes}</button>
                    </div>
                    <div class="reply">
                        ${isOwner ? '<button class="blue_empty_btn btn_update">Edit</button>': ''}
                        ${isLoggedIn ? '<button class="blue_empty_btn btn_reply">Reply</button>': ''}
                    </div>
                </div>
            </article>`;
            container.insertAdjacentHTML('beforeend', newEl);
    };
};
//to disable the form for comments if the user isn't logged in
function toDisableForm(isLoggedIn){
    if (!isLoggedIn){
        form.querySelector(".message_field").disabled = true;
        form.querySelector(".shadow_btn").disabled = true;
        form.querySelector(".main_btn").disabled = true;
    }
}
//formating of the date
function formatDate(isoString){
    const date = new Date(isoString);
    return date.toLocaleDateString(undefined, { day: 'numeric', month: 'long', year: 'numeric' });
};
//click function for creating a new comment
async function createComment(form, postID) { 
    const textValue = form .querySelector(".message_field") .value .trim(); 
    // don't send empty comment 
    if (textValue === "") { 
        return; 
    } 
    const data = { 
        text: textValue, 
        parent_comment_id: null 
    }; 
    try { 
        const response = await fetch( `http://localhost:8080/posts/${postID}/comments`, { 
            method: "POST",   
            headers: { "Content-Type": "application/json;charset=utf-8" }, 
            body: JSON.stringify(data) 
        } ); 
        if (!response.ok) { 
            throw new Error(`Error HTTP: ${response.status}`); 
        } 
        const result = await response.json(); 
        console.log("Successful send:", result); 
        // Make a form empty 
        form.querySelector(".message_field").value = ""; 
        // I could add render a comment here later
        // // renderNewComment(result); 
        } catch (error) { 
            console.error("Error creating comment:", error); 
    } 
}
// click events on update button
const postInside = document.querySelector(".post-inside");
postInside.addEventListener("click",  (event)=>{
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
    description.insertAdjacentHTML('afterend', elUpdate);
    const cancelBtn =comment.querySelector(".update-form .cancel")
    const submitBtn =comment.querySelector(".update-form .submit")
    const updateForm = comment.querySelector(".update-form");
    cancelBtn.addEventListener("click", ()=>{
        footer.style.display = "flex";
        description.style.display = "block";
        updateForm.remove();
    });
    submitBtn.addEventListener("click", ()=>{
        const value = comment.querySelector(".update-form .message_field").value;
        const trimmed = value.trim();
        if (trimmed === "") return;
        description.textContent = trimmed;
        footer.style.display = "flex";
        description.style.display = "block";
        updateForm.remove();
    });
});
