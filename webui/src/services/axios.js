import axios from "axios";

const instance = axios.create({
	baseURL: "/",
	timeout: 1000 * 5
});

export default instance;
