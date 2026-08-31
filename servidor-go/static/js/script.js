document.getElementById("imgClick").addEventListener("click", () => {
    if (document.body.style.background === "red") {
        document.body.style.background = "white";
    } else {
        document.body.style.background = "red";
    }
})