// function for parsing a page of all users
// `<a class="user" href="profile.html?id=${id}" aria-label="link to the ${username} profile">
//     <div class="profile">
//         <div class="image-container">
//             <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${username}">
//         </div>
//     </div>
//     <p class="user_name">${username}</p>
// <p>${description}</p>
// </a>`
// you need to add button update only for owner of user profile


const usersContainer = document.querySelector("#users");

async function getUsers() {
  try {
    const response = await fetch("http://localhost:8080/users");
    
    if (!response.ok) {
      throw new Error(`Failed to fetch users: ${response.status}`);
    }

    const data = await response.json();
    renderUsers(data.users);

  } catch (error) {
    console.error("Error loading users:", error);

    usersContainer.textContent = "Failed to load users.";
  }
}

function renderUsers(users) {
  usersContainer.replaceChildren();

  for (const user of users) {
    const userLink = document.createElement("a");
    userLink.className = "user";
    userLink.href = `profile.html?id=${encodeURIComponent(user.id)}`;
    userLink.setAttribute(
      "aria-label",
      `Link to the ${user.username} profile`
    );

    const profile = document.createElement("div");
    profile.className = "profile";

    const imageContainer = document.createElement("div");
    imageContainer.className = "image-container";

    const image = document.createElement("img");
    image.className = "author_img";
    image.src = "/src/static/imgs/default.jpg";
    // when support of files will be there
    // image.src = user.profilepic
    //     ? `/src/static/imgs/${user.profilepic}`
    //     : "/src/static/imgs/default.jpg";
    image.alt = `Photo of ${user.username}`;

    imageContainer.append(image);
    profile.append(imageContainer);

    const username = document.createElement("p");
    username.className = "user_name";
    username.textContent = `@${user.username}`;

    userLink.append(profile, username);

    if (user.description) {
      const description = document.createElement("p");
      description.textContent = user.description;
      userLink.append(description);
    }

    usersContainer.append(userLink);
  }
}

getUsers();
