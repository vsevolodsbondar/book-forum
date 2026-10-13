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
