import axios from "axios";

const backendUrl = `${window.location.protocol}//${window.location.hostname}:3000`;

const instance = axios.create({
	baseURL: backendUrl,
	timeout: 1000 * 5
});

export default instance;
