// login.js
// programming for login page

window.onload = () => {
	let submit_button = document.getElementById("login-submit");

	submit_button.onclick = () => {
		let ID = document.getElementById("login-id").value;
		window.location.href = `/?id=${ID}`; // go to index
	};
}
