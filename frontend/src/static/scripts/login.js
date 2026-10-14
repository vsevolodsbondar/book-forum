//change of file name when image was chosen from computer
const input = document.getElementById("file-field");
const label = document.querySelector(".file-field_label");

input.addEventListener("change", function () {
  if (this.files && this.files.length > 0) {
    label.textContent = this.files[0].name;
  }
});
//click on register and login button
const regBtn = document.getElementById("toRegister");
const loginBtn = document.getElementById("toLogin");
const formContainer = document.querySelector(".form-container");
regBtn.addEventListener("click", () => {
  formContainer.style.justifyContent = "flex-end";
});
loginBtn.addEventListener("click", () => {
  formContainer.style.justifyContent = "flex-start";
});

//add event listener for login submit form
const loginForm = document.getElementById("login");
// shows (or clears) the error text inside a form
function showFormError(form, text) {
  form.querySelector(".form_error").textContent = text;
}
loginForm.addEventListener("submit", async (event) => {
  showFormError(loginForm, "");
  // stop the browser from reloading the page
  event.preventDefault();
  const email = loginForm.elements.email.value;
  const password = loginForm.elements.password.value;
  try {
    const response = await fetch("http://localhost:8080/login", {
      method: "POST",
      credentials: "include", // lets the browser save the session cookie
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!response.ok) {
      const error = new Error(`Error HTTP: ${response.status}`);
      error.status = response.status; // keep the code to choose the message later
      throw error;
    }
    //if logged in redirect a user to index.html
    window.location.href = "index.html";
  } catch (error) {
    console.log("login failed", error);
    if (error instanceof TypeError) {
      showFormError(loginForm, "The server is not available. Try again later.");
    } else if (error.status === 401) {
      showFormError(loginForm, "Wrong email or password.");
    } else {
      showFormError(loginForm, "Something went wrong. Try again later.");
    }
  }
});

//registration form
const regForm = document.getElementById("reg");
regForm.addEventListener("submit", async (event) => {
  showFormError(regForm, "");
  event.preventDefault();
  const email = regForm.elements.email.value;
  const password = regForm.elements.password.value;
  const repeatPassword = regForm.elements.repeatPassword.value;
  const username = regForm.elements.username.value;
  const description = regForm.elements.description.value;
  //cheking if the passwords that were written 2 times are the same
  if (password != repeatPassword) {
    showFormError(regForm, "Passwords do not match.");
    return;
  }
  try {
    const response = await fetch("http://localhost:8080/users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, username, description }),
    });
    if (!response.ok) {
      const error = new Error(`Error HTTP: ${response.status}`);
      error.status = response.status; // keep the code to choose the message later
      throw error;
    }
    // registered -> show the login form and tell the user what to do next
    formContainer.style.justifyContent = "flex-start";
    showFormError(loginForm, "You are registered. Now you can log in.");
    loginForm.elements.email.value = email;
  } catch (error) {
    console.log("registration failed", error);
    if (error instanceof TypeError) {
      // fetch itself failed: the server is not reachable
      showFormError(regForm, "The server is not available. Try again later.");
    } else if (error.status === 409) {
      showFormError(regForm, "This email or username is already in use.");
    } else if (error.status === 400) {
      showFormError(
        regForm,
        "Check your data: password must have at least 15 characters, username 3-30 characters (letters, digits and _).",
      );
    } else {
      showFormError(regForm, "Something went wrong. Try again later.");
    }
  }
});
