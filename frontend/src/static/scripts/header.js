const header = document.querySelector("header.header");
const container = document.querySelector(".header .header-btn");
function renderLoggedOutHeader(){
    container.innerHTML = `
            <a href="login.html#reg" class="main_btn header_register">Register</a>
            <a href="login.html#login" class="shadow_btn header_login">Login</a>`;
};
function renderLoggedInHeader(){
   container.ineerHTML = `<a href="createPost.html" aria-label="link to the page create post" class="main_btn">Write a post</a>
            <div class="image-container">
                <img class="author_img" src="./src/static/imgs/photo.jpeg" alt="photo of user">
            </div>
            `;
};
renderLoggedOutHeader();