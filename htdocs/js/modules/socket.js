// NOTE: change this in production!
var URL_ROOT = "localhost:8080";

function connectWS(ID, func_recv) {
	// calling function should run connectWS in a try-catch block
	window.pacmacro_ws = new WebSocket(`ws://${URL_ROOT}/api/ws/${ID}`);

	if (window.pacmacro_ws === undefined)
		return false;

	// received API message
	window.pacmacro_ws.addEventListener("message", func_recv);

	return true;
}

export { connectWS };
