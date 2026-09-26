export async function getUserID() {
	try {
		const response = await fetch("http://localhost:8080/me", {
			credentials: "include",
		});
		if (!response.ok) {
			return null;
		}
		const userID = await response.json();
		return userID;
	} catch (error) {
		console.log("Error getting user ID:", error);
		return null;
	}
}
//
async function checkIsUserLoggedIn() {
	try {
		const userID = await getUserID();
		if (userID === null) {
			renderLoggedOutHeader();
			return;
		}
		const responseUser = await fetch(`http://localhost:8080/users/${userID}`);
		if (!responseUser.ok) {
			throw new Error(`Error HTTP: ${responseUser.status}`);
		}
		const userData = await responseUser.json();
		console.log(userData);
		renderLoggedInHeader(userData);
		renderLoggedInNav(userData);
	} catch (error) {
		console.log("the error was catched", error);
	}
}
checkIsUserLoggedIn();
//render of header
const container = document.querySelector(".header .header-btn");
function renderLoggedOutHeader() {
	container.innerHTML = `
            <a href="login.html#reg" class="main_btn header_register">Register</a>
            <a href="login.html#login" class="shadow_btn header_login">Login</a>`;
}
function renderLoggedInHeader(user) {
	container.innerHTML = `<a href="createPost.html" aria-label="link to the page create post" class="main_btn">Write a post</a>
            <a href="profile.html?id=${user.id}" class="image-container" aria-label="link to the your personal page">
                <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${user.username}">
            </a>
            <button class="blue_empty_btn">Log out</button>
            `;
}
//render navigation
const list = document.querySelector(".nav-list");
//renders navigation, check if there is already 4 li it's not gonna fill new elements
function renderLoggedInNav(user) {
	if (list.children.length !== 4) {
		list.insertAdjacentHTML(
			"beforeend",
			`<li><a class="nav-list_item" href="createPost.html">Create post</a></li>
        <li><a class="nav-list_item" href="myPosts.html?id=${user.id}">My posts</a></li>`,
		);
	}
}
