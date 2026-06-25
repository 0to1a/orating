import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
	input: '../openapi.json',
	output: {
		path: 'src/lib/api',
		format: 'prettier',
		lint: 'eslint'
	},
	plugins: [
		'@hey-api/client-fetch',
		{
			name: '@hey-api/sdk',
			operationId: true
		},
		'@hey-api/typescript'
	]
});
