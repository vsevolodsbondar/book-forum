//
async function checkIsUserLoggedIn(){
        try{
        const response = await fetch(`http://localhost:8080/me`);
        if (!response.ok){
            renderLoggedOutHeader();
            return;
        };
        const userID = await response.json();
        const responseUser = await fetch(`http://localhost:8080/users/${userID}`);
        if (!responseUser.ok){
            throw new Error(`Error HTTP: ${response.status}`); 
        };
        const userData = await response.json();
        renderLoggedInHeader(userData);
        renderLoggedInNav(userData);
    } catch(error){
        console.log("the error was catched", error);
    }
};
checkIsUserLoggedIn();
//render of header
const header = document.querySelector("header.header");
const container = document.querySelector(".header .header-btn");
function renderLoggedOutHeader(){
    container.innerHTML = `
            <a href="login.html#reg" class="main_btn header_register">Register</a>
            <a href="login.html#login" class="shadow_btn header_login">Login</a>`;
};
function renderLoggedInHeader(user){
   container.innerHTML = `<a href="createPost.html" aria-label="link to the page create post" class="main_btn">Write a post</a>
            <div class="image-container">
                <img class="author_img" src="./src/static/imgs/default.jpg" alt="photo of ${user.username}">
            </div>
            `;
};
//render navigation
const list = document.querySelector(".nav-list")
function renderLoggedInNav(user){
    let newEl;
    newEl.innerHTML = `<li><a class="nav-list_item" href="createPost.html">Create post</a></li>
    <li><a class="nav-list_item" href="myPosts?id=${user.id}.html">My posts</a></li>`; 
    list.insertAdjacentHTML("beforeend", newEl);
}