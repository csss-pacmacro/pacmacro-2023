// login.js
// programming for login page

import { ribbons, getID } from "./pacmacro.js";

window.onload = () => {
	ribbons();

	let ID = document.getElementById("login-id");
	ID.value = getID();
	let submit_button = document.getElementById("login-submit");

	submit_button.onclick = () => {
		window.location.href = `/?id=${ID.value}`; // go to index
	};
}
