import { client } from './api/client.gen';
import { handleUnauthorized, getToken } from './auth';

// Configure the generated hey-api client
client.setConfig({ baseUrl: '' });

// Attach auth token to every request
client.interceptors.request.use((request) => {
	const token = getToken();
	if (token) {
		request.headers.set('Authorization', `Bearer ${token}`);
	}
	return request;
});

// Handle 401 globally
client.interceptors.response.use(async (response) => {
	if (response.status === 401) {
		await handleUnauthorized();
	}
	return response;
});

// Re-export for convenience
export { client as apiClient };
export * from './api/index';
