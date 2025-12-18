import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Download, Loader, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import axios from 'axios';

const API_BASE = 'http://localhost:8080/api/v1';

const SyncManga = () => {
  const [query, setQuery] = useState('');
  const [limit, setLimit] = useState(10);
  const [syncing, setSyncing] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);

  const handleSync = async (e) => {
    e.preventDefault();
    
    if (!query.trim()) {
      return;
    }

    setSyncing(true);
    setError(null);
    setResult(null);

    try {
      const response = await axios.post(`${API_BASE}/manga/sync`, {
        query: query,
        limit: parseInt(limit)
      });

      setResult(response.data);
    } catch (err) {
      console.error('Sync error:', err);
      setError(err.response?.data?.error || err.message);
    } finally {
      setSyncing(false);
    }
  };

  const popularQueries = [
    'One Piece',
    'Naruto', 
    'Attack on Titan',
    'Death Note',
    'Dragon Ball',
    'My Hero Academia',
    'Demon Slayer',
    'Tokyo Ghoul'
  ];

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4 max-w-4xl">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8"
        >
          <div className="flex items-center gap-3 mb-4">
            <Download className="w-8 h-8 text-blue-600" />
            <h1 className="text-4xl font-bold text-gray-900">Sync Manga from MAL</h1>
          </div>
          <p className="text-gray-600">
            Fetch manga from MyAnimeList and store only those with readable chapters on MangaDex
          </p>
        </motion.div>

        {/* Info Box */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6"
        >
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-blue-900 mb-1">How it works:</h3>
              <ol className="text-sm text-blue-800 space-y-1 list-decimal list-inside">
                <li>Fetches manga from MyAnimeList based on your search query</li>
                <li>Checks if each manga has chapters available on MangaDex</li>
                <li>Only stores manga with readable chapters in the local database</li>
                <li>Synced manga will be searchable via gRPC search</li>
              </ol>
            </div>
          </div>
        </motion.div>

        {/* Sync Form */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="bg-white rounded-lg shadow-md p-6 mb-6"
        >
          <form onSubmit={handleSync} className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Search Query
              </label>
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="e.g., One Piece, Naruto..."
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={syncing}
              />
            </div>

            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Limit (max 50)
              </label>
              <input
                type="number"
                value={limit}
                onChange={(e) => setLimit(Math.min(50, Math.max(1, parseInt(e.target.value) || 10)))}
                min="1"
                max="50"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={syncing}
              />
            </div>

            <button
              type="submit"
              disabled={!query.trim() || syncing}
              className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition font-semibold disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {syncing ? (
                <>
                  <Loader className="w-5 h-5 animate-spin" />
                  <span>Syncing...</span>
                </>
              ) : (
                <>
                  <Download className="w-5 h-5" />
                  <span>Start Sync</span>
                </>
              )}
            </button>
          </form>

          {/* Quick Suggestions */}
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-sm text-gray-600 mb-3">Quick sync:</p>
            <div className="flex flex-wrap gap-2">
              {popularQueries.map((q) => (
                <button
                  key={q}
                  onClick={() => setQuery(q)}
                  className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm hover:bg-gray-200 transition"
                  disabled={syncing}
                >
                  {q}
                </button>
              ))}
            </div>
          </div>
        </motion.div>

        {/* Error Display */}
        {error && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6"
          >
            <div className="flex items-start gap-3">
              <XCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
              <div>
                <h3 className="font-semibold text-red-900 mb-1">Sync Failed</h3>
                <p className="text-sm text-red-800">{error}</p>
              </div>
            </div>
          </motion.div>
        )}

        {/* Success Result */}
        {result && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-white rounded-lg shadow-md p-6"
          >
            <div className="flex items-center gap-3 mb-4">
              <CheckCircle className="w-6 h-6 text-green-600" />
              <h2 className="text-xl font-bold text-gray-900">Sync Completed</h2>
            </div>

            {/* Summary Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <div className="bg-blue-50 rounded-lg p-4">
                <div className="text-2xl font-bold text-blue-900">{result.total_fetched}</div>
                <div className="text-sm text-blue-600">Fetched from MAL</div>
              </div>
              <div className="bg-green-50 rounded-lg p-4">
                <div className="text-2xl font-bold text-green-900">{result.synced}</div>
                <div className="text-sm text-green-600">Successfully Synced</div>
              </div>
              <div className="bg-yellow-50 rounded-lg p-4">
                <div className="text-2xl font-bold text-yellow-900">{result.skipped}</div>
                <div className="text-sm text-yellow-600">Skipped (No Chapters)</div>
              </div>
              <div className="bg-red-50 rounded-lg p-4">
                <div className="text-2xl font-bold text-red-900">{result.failed}</div>
                <div className="text-sm text-red-600">Failed</div>
              </div>
            </div>

            {/* Details */}
            {result.details && result.details.length > 0 && (
              <div>
                <h3 className="font-semibold text-gray-900 mb-3">Details:</h3>
                <div className="space-y-2 max-h-96 overflow-y-auto">
                  {result.details.map((detail, index) => (
                    <div
                      key={index}
                      className={`p-3 rounded-lg text-sm ${
                        detail.startsWith('✅')
                          ? 'bg-green-50 text-green-800'
                          : 'bg-gray-50 text-gray-800'
                      }`}
                    >
                      {detail}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Success Message */}
            <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-900 font-medium">{result.message}</p>
              <p className="text-sm text-green-700 mt-1">
                Synced manga are now searchable via gRPC search in the Search page!
              </p>
            </div>
          </motion.div>
        )}
      </div>
    </div>
  );
};

export default SyncManga;
