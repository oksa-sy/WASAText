import axios from "axios";

// calcolo automatico del protocollo e dell'hostname su porta 3000
const backendUrl = `${window.location.protocol}//${window.location.hostname}:3000`;

const instance = axios.create({
	baseURL: backendUrl,
	timeout: 1000 * 5
});

export default instance;
