import { watchLocation } from "./modules/location.js";
import { connectWS } from "./modules/socket.js";

window.onload = async () => {
	/* LOCATION */

	let location_button = document.getElementById("location-button");
	let location_p = document.getElementById("location");

	// wait for user to prompt watching
	location_button.onclick = () => {
		// watch location and store information in p
		watchLocation(location_p);
	};

	/* REGISTRATION */

	let register_button = document.getElementById("register-button");
	let register_p = document.getElementById("register");

	register_button.onclick = async () => {
		const form = new FormData();

		form.append("type", "2");    // TypeAdmin
		form.append("pass", "1234"); // adminPassword

		let resp;

		try {
			resp = await fetch("/api/player/register", {
				method: "POST",
				body: form
			});
		} catch {
			register_p.innerHTML = "There was an error reaching the server.";
		}

		if (resp.ok) {
			let ID = await resp.text();

			// store ID in cookie
			document.cookie = `id=${ID}`;

			register_p.innerHTML = `Registered Admin with ID ${ID}.`;
		} else
			register_p.innerHTML = "Status code not OK.";
	};

	/* WEBSOCKETS */

	let socket_button = document.getElementById("socket-button");
	let sclose_button = document.getElementById("socket-close-button");
	let socket_p = document.getElementById("socket");

	socket_button.onclick = () => {
		let ID = "";

		let cookies = document.cookie;
		cookies = cookies.split(';').map(v => v.split('='));
		for (const c of cookies) {
			if (c[0].trim() == "id")
				ID = c[1];
		}

		if (ID.length == 0) {
			socket_p.innerHTML = "Player is not registered.";
			return;
		}

		try {
			connectWS(ID, (e) => {
				socket_p.innerHTML = e.data;
			});
		} catch {
			socket_p.innerHTML = "Failed to connect to WS.";
			return;
		}

		sclose_button.style.visibility = "visible";
	};

	sclose_button.style.visibility = "hidden";

	sclose_button.onclick = () => {
		if (window.pacmacro_ws) {
			window.pacmacro_ws.close();
			window.pacmacro_ws = undefined;
		}
		sclose_button.style.visibility = "hidden";
	}
};
