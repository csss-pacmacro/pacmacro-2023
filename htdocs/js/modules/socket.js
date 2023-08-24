// NOTE: change this in production!
var URL_ROOT = "localhost:8080";

function initWS() {
	try {
		window.pacmacro_ws = new WebSocket(`ws://${URL_ROOT}/api/ws/`);
	} catch {
		return;
	}

	// received API message
	window.pacmacro_ws.addEventListener("message", (e) => {
		// log API message
		console.log(e.data);
	});
}

export { initWS };
