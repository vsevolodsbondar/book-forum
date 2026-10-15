import { getUserID } from "./header.js";

const section = document.querySelector(".user-profile");
const titleEl = section.querySelector(".user_title");
const nameEl = section.querySelector(".user_name");
const descEl = section.querySelector(".description");
const avatarEl = section.querySelector(".author_img");

async function getUser() {
  try {
    // ?id=... in the address, or the logged in user if there is no id
    const params = new URLSearchParams(window.location.search);
    const userID = params.get("id") ?? (await getUserID());
    if (userID === null) {
      section.innerHTML = "<p>Log in to see your profile</p>";
      return;
    }

    const response = await fetch(`http://localhost:8080/users/${userID}`);
    if (response.status === 404) {
      section.innerHTML = "<p>User not found</p>";
      return;
    }
    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }
    const user = await response.json();

    // textContent (not innerHTML), so the text of the user can't inject HTML
    titleEl.textContent = `@${user.username}`;
    nameEl.textContent = user.name;
    descEl.textContent = user.description;
    avatarEl.src = "/src/static/imgs/default.jpg";
    avatarEl.alt = `photo of ${user.username}`;
  } catch (error) {
    console.log("the error was catched", error);
    section.innerHTML = "<p>Could not load the profile. Try again later.</p>";
  }
}
getUser();
