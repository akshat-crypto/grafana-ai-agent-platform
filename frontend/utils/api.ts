import axios, { AxiosInstance, AxiosResponse } from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Add response interceptor to handle auth errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Auth endpoints
  async register(data: {
    email: string;
    password: string;
    firstName: string;
    lastName: string;
  }) {
    return this.client.post('/api/auth/register', {
      email: data.email,
      password: data.password,
      first_name: data.firstName,
      last_name: data.lastName,
    });
  }

  async login(data: { email: string; password: string }) {
    return this.client.post('/api/auth/login', data);
  }

  async logout() {
    return this.client.post('/api/auth/logout');
  }

  async getProfile() {
    return this.client.get('/api/profile');
  }

  // Kubernetes endpoints
  async validateCluster(data: { kube_config: string }) {
    return this.client.post('/api/kubernetes/validate', {kube_config: data.kube_config});
  }

  async addCluster(data: { name: string; kube_config: string }) {
    return this.client.post('/api/kubernetes/clusters', {name: data.name, kube_config: data.kube_config});
  }

  async getClusters() {
    return this.client.get('/api/kubernetes/clusters');
  }

  async deleteCluster(id: string) {
    return this.client.delete(`/api/kubernetes/clusters/${id}`);
  }

  async refreshClusterStatus(id: string) {
    return this.client.post(`/api/kubernetes/clusters/${id}/refresh`);
  }

  async getClusterResources(id: string) {
    return this.client.get(`/api/kubernetes/clusters/${id}/resources`);
  }

  // AI Agent endpoints
  async queryAgent(data: { query: string; clusterId?: number }) {
    return this.client.post('/api/agent/query', data);
  }

  async deployStack(data: {
    stackName: string;
    clusterId: number;
    query: string;
  }) {
    return this.client.post('/api/agent/deploy', data);
  }

  async getQueryHistory() {
    return this.client.get('/api/agent/queries');
  }

  async getDeploymentHistory() {
    return this.client.get('/api/agent/deployments');
  }

  // Health check
  async healthCheck() {
    return this.client.get('/health');
  }
}

export const apiClient = new ApiClient();
export default apiClient; 