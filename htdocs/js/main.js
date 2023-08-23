import { watchLocation } from "./modules/location.js";

window.onload = () => {
	let p = document.getElementById("location");
	let button = document.getElementById("location-button");

	// wait for user to prompt watching
	button.onclick = () => {
		// watch location and store information in p
		watchLocation(p);
	};
};
