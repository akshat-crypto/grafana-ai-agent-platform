import React, { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { PaperAirplaneIcon, PlusIcon, ChartBarIcon, CogIcon, SparklesIcon } from '@heroicons/react/24/outline';
import apiClient from '../utils/api';
import toast from 'react-hot-toast';

interface Cluster {
  id: number;
  name: string;
  status: string;
  isActive: boolean;
  version: string;
}

const Dashboard: React.FC = () => {
  const [query, setQuery] = useState('');
  const [response, setResponse] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<number | undefined>();

  useEffect(() => {
    loadClusters();
  }, []);

  const loadClusters = async () => {
    try {
      const response = await apiClient.getClusters();
      setClusters(response?.data || []);
    } catch (error) {
      console.error('Failed to load clusters:', error);
      setClusters([]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setIsLoading(true);
    try {
      const response = await apiClient.queryAgent({
        query: query.trim(),
        clusterId: selectedCluster,
      });
      
      setResponse(response.data.response);
      setQuery('');
      toast.success('Query processed successfully!');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to process query');
    } finally {
      setIsLoading(false);
    }
  };

  const quickQueries = [
    { text: 'Install Grafana and Prometheus stack', icon: ChartBarIcon },
    { text: 'Deploy ELK stack for logging', icon: SparklesIcon },
    { text: 'Set up monitoring with Prometheus', icon: ChartBarIcon },
    { text: 'Create a basic nginx deployment', icon: CogIcon },
    { text: 'Install Istio service mesh', icon: SparklesIcon },
  ];

  return (
    <Layout>
      <div className="space-y-8">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl p-8 text-white">
          <h1 className="text-3xl font-bold mb-2">Welcome to K8s AI Platform</h1>
          <p className="text-blue-100 text-lg">
            Manage your Kubernetes clusters with AI-powered assistance
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl transition-shadow duration-300">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active Clusters</p>
                <p className="text-2xl font-bold text-gray-900">{clusters.filter(c => c.isActive).length}</p>
              </div>
              <div className="h-12 w-12 bg-blue-100 rounded-lg flex items-center justify-center">
                <ChartBarIcon className="h-6 w-6 text-blue-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl transition-shadow duration-300">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Clusters</p>
                <p className="text-2xl font-bold text-gray-900">{clusters.length}</p>
              </div>
              <div className="h-12 w-12 bg-green-100 rounded-lg flex items-center justify-center">
                <CogIcon className="h-6 w-6 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl transition-shadow duration-300">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">AI Queries</p>
                <p className="text-2xl font-bold text-gray-900">∞</p>
              </div>
              <div className="h-12 w-12 bg-purple-100 rounded-lg flex items-center justify-center">
                <SparklesIcon className="h-6 w-6 text-purple-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* AI Chat Interface */}
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-xl font-semibold text-gray-900 flex items-center">
                <SparklesIcon className="h-6 w-6 text-purple-600 mr-2" />
                AI Assistant
              </h3>
              <div className="flex items-center space-x-2">
                <div className="h-2 w-2 bg-green-500 rounded-full animate-pulse"></div>
                <span className="text-sm text-gray-500">Online</span>
              </div>
            </div>
            
            {/* Cluster Selection */}
            {clusters.length > 0 && (
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Select Cluster (Optional)
                </label>
                <select
                  value={selectedCluster || ''}
                  onChange={(e) => setSelectedCluster(e.target.value ? Number(e.target.value) : undefined)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                  <option value="">No specific cluster</option>
                  {clusters.map(cluster => (
                    <option key={cluster.id} value={cluster.id}>
                      {cluster.name} (v{cluster.version})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Quick Queries */}
            <div className="mb-6">
              <p className="text-sm font-medium text-gray-700 mb-3">Quick Queries:</p>
              <div className="flex flex-wrap gap-2">
                {quickQueries.map((quickQuery, index) => (
                  <button
                    key={index}
                    onClick={() => setQuery(quickQuery.text)}
                    className="inline-flex items-center px-3 py-1.5 text-sm bg-gradient-to-r from-blue-50 to-purple-50 hover:from-blue-100 hover:to-purple-100 text-gray-700 rounded-full transition-all duration-200 border border-gray-200 hover:border-gray-300"
                  >
                    <quickQuery.icon className="h-4 w-4 mr-1.5" />
                    {quickQuery.text}
                  </button>
                ))}
              </div>
            </div>

            {/* Chat Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Ask about Kubernetes operations
                </label>
                <textarea
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="e.g., Install Grafana and Prometheus stack, or help me debug a deployment..."
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent resize-none h-24"
                  disabled={isLoading}
                />
              </div>
              
              <button
                type="submit"
                disabled={isLoading || !query.trim()}
                className="w-full flex items-center justify-center px-4 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-medium rounded-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <PaperAirplaneIcon className="h-4 w-4 mr-2" />
                {isLoading ? 'Processing...' : 'Send Query'}
              </button>
            </form>

            {/* Response */}
            {response && (
              <div className="mt-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
                <h4 className="font-medium text-gray-900 mb-3 flex items-center">
                  <SparklesIcon className="h-5 w-5 text-purple-600 mr-2" />
                  AI Response:
                </h4>
                <div className="prose prose-sm max-w-none">
                  <pre className="whitespace-pre-wrap text-sm text-gray-700 bg-white p-4 rounded border overflow-auto max-h-96">
                    {response}
                  </pre>
                </div>
              </div>
            )}
          </div>

          {/* Clusters Overview */}
          <div className="space-y-6">
            {/* Active Clusters */}
            <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-xl font-semibold text-gray-900 flex items-center">
                  <ChartBarIcon className="h-6 w-6 text-blue-600 mr-2" />
                  Active Clusters
                </h3>
                <button className="inline-flex items-center px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors duration-200">
                  <PlusIcon className="h-4 w-4 mr-1.5" />
                  Add Cluster
                </button>
              </div>
              
              <div className="space-y-3">
                {clusters.filter(c => c.isActive).map(cluster => (
                  <div key={cluster.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-200 hover:bg-gray-100 transition-colors duration-200">
                    <div>
                      <p className="font-medium text-gray-900">{cluster.name}</p>
                      <p className="text-sm text-gray-500">v{cluster.version}</p>
                    </div>
                    <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded-full">
                      Active
                    </span>
                  </div>
                ))}
                {clusters.filter(c => c.isActive).length === 0 && (
                  <div className="text-center py-8">
                    <CogIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                    <p className="text-gray-500 text-sm">No active clusters</p>
                    <p className="text-gray-400 text-xs mt-1">Add your first cluster to get started</p>
                  </div>
                )}
              </div>
            </div>

            {/* Quick Actions */}
            <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
              <h3 className="text-xl font-semibold text-gray-900 mb-4 flex items-center">
                <SparklesIcon className="h-6 w-6 text-purple-600 mr-2" />
                Quick Actions
              </h3>
              <div className="space-y-3">
                <button className="w-full text-left p-4 bg-gradient-to-r from-blue-50 to-blue-100 hover:from-blue-100 hover:to-blue-200 rounded-lg transition-all duration-200 border border-blue-200">
                  <p className="font-medium text-blue-900">Add New Cluster</p>
                  <p className="text-sm text-blue-700">Connect a Kubernetes cluster</p>
                </button>
                <button className="w-full text-left p-4 bg-gradient-to-r from-green-50 to-green-100 hover:from-green-100 hover:to-green-200 rounded-lg transition-all duration-200 border border-green-200">
                  <p className="font-medium text-green-900">Deploy Stack</p>
                  <p className="text-sm text-green-700">Install monitoring or logging</p>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default Dashboard; 