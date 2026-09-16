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
