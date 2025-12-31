import {fileURLToPath, URL} from 'node:url'

import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig(({command, mode, ssrBuild}) => {
	const ret = {
		plugins: [vue()],
		resolve: {
			alias: {
				'@': fileURLToPath(new URL('./src', import.meta.url))
			}
		},
		// proxy per far comunicare backend e frontend senza errori CORS
		server: {
            port: 5173,
            proxy: {
                '^/session': { target: 'http://localhost:3000' },
                '^/users': { target: 'http://localhost:3000' },
                '^/conversations': { target: 'http://localhost:3000' },
                '^/groups': { target: 'http://localhost:3000' },
                '^/messages': { target: 'http://localhost:3000' },
            }
        }
	};
	ret.define = {
		// Do not modify this constant, it is used in the evaluation.
		"__API_URL__": JSON.stringify("http://localhost:3000"),
	};
	return ret;
})
