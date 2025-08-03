import React, { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { PlusIcon, TrashIcon, CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import apiClient from '../utils/api';
import toast from 'react-hot-toast';

interface Cluster {
  id: number;
  name: string;
  status: string;
  isActive: boolean;
  version: string;
}

interface ValidationResult {
  is_valid: boolean;
  version?: string;
  error?: string;
  server_url?: string;
}

const KubernetesPage: React.FC = () => {
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [isAddingCluster, setIsAddingCluster] = useState(false);
  const [clusterName, setClusterName] = useState('');
  const [kubeConfig, setKubeConfig] = useState('');
  const [isValidating, setIsValidating] = useState(false);
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null);

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

  const validateKubeConfig = async () => {
    if (!kubeConfig.trim()) {
      toast.error('Please enter a kubeconfig');
      return;
    }

    setIsValidating(true);
    try {
      const response = await apiClient.validateCluster({ kubeConfig: kubeConfig.trim() });
      setValidationResult(response.data);
      
      if (response.data.is_valid) {
        toast.success('Kubeconfig is valid!');
      } else {
        toast.error('Invalid kubeconfig');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Validation failed');
    } finally {
      setIsValidating(false);
    }
  };

  const addCluster = async () => {
    if (!clusterName.trim() || !kubeConfig.trim()) {
      toast.error('Please fill in all fields');
      return;
    }

    if (!validationResult?.is_valid) {
      toast.error('Please validate the kubeconfig first');
      return;
    }

    try {
      await apiClient.addCluster({
        name: clusterName.trim(),
        kubeConfig: kubeConfig.trim(),
      });
      
      toast.success('Cluster added successfully!');
      setIsAddingCluster(false);
      setClusterName('');
      setKubeConfig('');
      setValidationResult(null);
      loadClusters();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to add cluster');
    }
  };

  const deleteCluster = async (id: number) => {
    if (!confirm('Are you sure you want to delete this cluster?')) return;

    try {
      await apiClient.deleteCluster(id.toString());
      toast.success('Cluster deleted successfully!');
      loadClusters();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to delete cluster');
    }
  };

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Kubernetes Clusters</h1>
            <p className="text-gray-600">Manage your Kubernetes cluster connections</p>
          </div>
          <button
            onClick={() => setIsAddingCluster(true)}
            className="btn-primary flex items-center space-x-2"
          >
            <PlusIcon className="h-4 w-4" />
            <span>Add Cluster</span>
          </button>
        </div>

        {/* Add Cluster Modal */}
        {isAddingCluster && (
          <div className="card">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Add New Cluster</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Cluster Name
                </label>
                <input
                  type="text"
                  value={clusterName}
                  onChange={(e) => setClusterName(e.target.value)}
                  placeholder="Enter cluster name"
                  className="input-field"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Kubeconfig
                </label>
                <textarea
                  value={kubeConfig}
                  onChange={(e) => setKubeConfig(e.target.value)}
                  placeholder="Paste your kubeconfig here..."
                  className="input-field h-32 resize-none font-mono text-sm"
                />
              </div>

              {/* Validation Result */}
              {validationResult && (
                <div className={`p-3 rounded-lg ${
                  validationResult.is_valid 
                    ? 'bg-green-50 border border-green-200' 
                    : 'bg-red-50 border border-red-200'
                }`}>
                  <div className="flex items-center space-x-2">
                    {validationResult.is_valid ? (
                      <CheckCircleIcon className="h-5 w-5 text-green-500" />
                    ) : (
                      <XCircleIcon className="h-5 w-5 text-red-500" />
                    )}
                    <span className={`font-medium ${
                      validationResult.is_valid ? 'text-green-800' : 'text-red-800'
                    }`}>
                      {validationResult.is_valid ? 'Valid' : 'Invalid'} Configuration
                    </span>
                  </div>
                  {validationResult.version && (
                    <p className="text-sm text-gray-600 mt-1">
                      Version: {validationResult.version}
                    </p>
                  )}
                  {validationResult.error && (
                    <p className="text-sm text-red-600 mt-1">
                      Error: {validationResult.error}
                    </p>
                  )}
                </div>
              )}

              <div className="flex space-x-3">
                <button
                  onClick={validateKubeConfig}
                  disabled={isValidating || !kubeConfig.trim()}
                  className="btn-secondary"
                >
                  {isValidating ? 'Validating...' : 'Validate'}
                </button>
                <button
                  onClick={addCluster}
                  disabled={!clusterName.trim() || !kubeConfig.trim() || !validationResult?.is_valid}
                  className="btn-primary"
                >
                  Add Cluster
                </button>
                <button
                  onClick={() => {
                    setIsAddingCluster(false);
                    setClusterName('');
                    setKubeConfig('');
                    setValidationResult(null);
                  }}
                  className="btn-secondary"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Clusters List */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {clusters.map(cluster => (
            <div key={cluster.id} className="card">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">{cluster.name}</h3>
                  <p className="text-sm text-gray-500">v{cluster.version}</p>
                </div>
                <button
                  onClick={() => deleteCluster(cluster.id)}
                  className="p-1 text-red-500 hover:text-red-700 hover:bg-red-50 rounded"
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">Status:</span>
                  <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                    cluster.isActive 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {cluster.isActive ? 'Active' : 'Inactive'}
                  </span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">Version:</span>
                  <span className="text-sm font-medium text-gray-900">{cluster.version}</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        {clusters.length === 0 && (
          <div className="text-center py-12">
            <div className="text-gray-400 mb-4">
              <svg className="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No clusters yet</h3>
            <p className="text-gray-500 mb-4">Get started by adding your first Kubernetes cluster</p>
            <button
              onClick={() => setIsAddingCluster(true)}
              className="btn-primary"
            >
              Add Your First Cluster
            </button>
          </div>
        )}
      </div>
    </Layout>
  );
};

export default KubernetesPage; 