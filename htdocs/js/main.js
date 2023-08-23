import { watchLocation } from "./modules/location.js";
import { initWS } from "./modules/socket.js";

window.onload = () => {
	let p = document.getElementById("location");
	let button = document.getElementById("location-button");

	// initialize websocket connection with API
	initWS();

	// wait for user to prompt watching
	button.onclick = () => {
		// watch location and store information in p
		watchLocation(p);
	};
};
