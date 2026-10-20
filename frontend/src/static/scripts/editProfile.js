import { getUserID } from "./header.js";

const form = document.querySelector("#editProfile");
const usernameInput = document.querySelector("#username");
const nameInput = document.querySelector("#name");
const descriptionInput = document.querySelector("#description");
const avatarInput = document.querySelector("#avatar");
const errorEl = document.querySelector(".form_error");

let originalUser = {};

async function loadProfile() {
    try {
        const userID = await getUserID();

        if (userID == null) {
            window.location.href = "login.html";
            return;
        }

        const response = await fetch(
            `http://localhost:8080/users/${userID}`,
            {
                credentials: "include"
            }
        );

        if (!response.ok) {
            throw new Error("Could not load your profile.");
        }

        const user = await response.json();

        // Save the original values for comparison
        originalUser = {
            username: user.username ?? "",
            name: user.name ?? "",
            description: user.description ?? ""
        };

        // Fill in the form
        usernameInput.value = originalUser.username;
        nameInput.value = originalUser.name;
        descriptionInput.value = originalUser.description;

    } catch (error) {
        console.error(error);
        errorEl.textContent = error.message;
    }
}

form.addEventListener("submit", async (event) => {
    event.preventDefault();
    errorEl.textContent = "";

    // Collect only fields that have changed
    const body = new FormData();

    const username = usernameInput.value.trim();
    const name = nameInput.value.trim();
    const description = descriptionInput.value.trim();

    if (username !== originalUser.username) {
        body.append("username", username);
    }

    if (name !== originalUser.name) {
        body.append("name", name);
    }

    if (description !== originalUser.description) {
        body.append("description", description);
    }

    if (avatarInput.files.length > 0) {
        body.append("avatar", avatarInput.files[0]);
    }

    // Nothing changed: don't send a request
    if ([...body.keys()].length === 0) {
        alert("You haven't changed anything!");
        return;
    }

    try {
        const userID = await getUserID();

        if (userID == null) {
            window.location.href = "login.html";
            return;
        }

        const response = await fetch(
            `http://localhost:8080/users/${userID}`,
            {
                method: "PATCH",
                credentials: "include",
                body
            }
        );

        if (!response.ok) {
            const error = await response.json();
            errorEl.textContent = error.message || "Something went wrong.";
            return;
        }

        originalUser = {
            username,
            name,
            description
        };

        alert("Profile updated!");
        window.location.href = `http://localhost:5173/profile.html?id=${userID}`;

    } catch (error) {
        console.error(error);
        errorEl.textContent = error.message;
    }
});

loadProfile();
