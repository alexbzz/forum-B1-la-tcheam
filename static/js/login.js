document.addEventListener("DOMContentLoaded", () => {
    const inputs = document.querySelectorAll(".input-group input");

    inputs.forEach(input => {
        input.addEventListener("focus", () => {
            input.parentElement.classList.add("active");
        });
        input.addEventListener("blur", () => {
            if (input.value === "") {
                input.parentElement.classList.remove("active");
            }
        });
    });

    // Fade-in effect
    const container = document.querySelector(".container");
    container.style.opacity = 0;
    setTimeout(() => {
        container.style.transition = "opacity 1s ease";
        container.style.opacity = 1;
    }, 200);
});
