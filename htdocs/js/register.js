// register.js
// programming for the Player Registration page (register.html)

import { URL_ROOT, saveID } from "./pacmacro.js";

window.onload = () => {
	let submit_button = document.getElementById("register-submit");

	submit_button.onclick = async () => {
		let type = document.getElementById("register-type").value;
		let pass = document.getElementById("register-pass").value;
		let stat = document.getElementById("register-status");

		let form = new FormData;
		form.append("type", type);
		form.append("pass", pass);

		let ID;

		try {
			ID = await fetch("/api/player/register", {
				method: "POST",
				body: form
			});
		} catch {
			stat.innerHTML = "Couldn't contact API.";
			return;
		}

		if (!ID.ok) {
			stat.innerHTML = `Error ${ID.status}`;
			return;
		}

		ID = await ID.text();
		saveID(ID); // save ID in cookies
		window.location.href = "/"; // go to root
	};
}
