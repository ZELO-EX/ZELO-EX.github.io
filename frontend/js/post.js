const buttons = document.querySelectorAll(".copy-code");
const codes = document.querySelectorAll("pre.src code");

buttons.forEach((item, index) => {
    item.addEventListener("click", (event) => {
	navigator.clipboard.writeText(codes[index].textContent);

	item.style.color="#0f0";
	setTimeout(() => {
	    item.style.color="#fff";
	}, 1000);
    })
})
