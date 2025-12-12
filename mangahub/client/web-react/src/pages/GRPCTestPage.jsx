import React, { useState, useEffect } from 'react';
import grpcService from '../services/grpcService';
import authService from '../services/authService';
import LoadingSpinner from '../components/LoadingSpinner';

const GRPCTestPage = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [grpcAvailable, setGrpcAvailable] = useState(null);
  const [error, setError] = useState(null);
  
  // UC-014: Get Manga
  const [getMangaId, setGetMangaId] = useState('one-piece');
  const [mangaResult, setMangaResult] = useState(null);
  
  // UC-015: Search Manga
  const [searchQuery, setSearchQuery] = useState('');
  const [searchLimit, setSearchLimit] = useState(10);
  const [searchResults, setSearchResults] = useState(null);
  
  // UC-016: Update Progress
  const [progressMangaId, setProgressMangaId] = useState('one-piece');
  const [progressChapter, setProgressChapter] = useState(1);
  const [progressStatus, setProgressStatus] = useState('reading');
  const [progressResult, setProgressResult] = useState(null);

  useEffect(() => {
    const checkAuth = () => {
      const authenticated = authService.isAuthenticated();
      setIsAuthenticated(authenticated);
    };

    const checkGRPCAvailability = async () => {
      if (authService.isAuthenticated()) {
        try {
          const available = await grpcService.isAvailable();
          setGrpcAvailable(available);
        } catch (err) {
          setGrpcAvailable(false);
        }
      }
    };

    checkAuth();
    checkGRPCAvailability();
  }, []);

  // UC-014: Get Manga via gRPC
  const handleGetManga = async () => {
    setLoading(true);
    setError(null);
    setMangaResult(null);
    
    try {
      const result = await grpcService.getManga(getMangaId);
      setMangaResult(result);
    } catch (err) {
      setError(`Get Manga Error: ${err.error || err.message || 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  // UC-015: Search Manga via gRPC
  const handleSearchManga = async () => {
    setLoading(true);
    setError(null);
    setSearchResults(null);
    
    try {
      const result = await grpcService.searchManga(searchQuery, searchLimit);
      setSearchResults(result);
    } catch (err) {
      setError(`Search Manga Error: ${err.error || err.message || 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  // UC-016: Update Progress via gRPC
  const handleUpdateProgress = async () => {
    setLoading(true);
    setError(null);
    setProgressResult(null);
    
    try {
      const result = await grpcService.updateProgress(
        progressMangaId,
        parseInt(progressChapter),
        progressStatus
      );
      setProgressResult(result);
    } catch (err) {
      setError(`Update Progress Error: ${err.error || err.message || 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 text-center">
          <h2 className="text-2xl font-bold text-yellow-800 mb-2">Authentication Required</h2>
          <p className="text-yellow-700">Please log in to test gRPC features.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-4xl font-bold mb-6">gRPC Service Test Page</h1>
      
      {/* Service Status */}
      <div className={`mb-6 p-4 rounded-lg ${grpcAvailable === null ? 'bg-gray-100' : grpcAvailable ? 'bg-green-100' : 'bg-red-100'}`}>
        <h2 className="text-xl font-semibold mb-2">
          gRPC Service Status: {grpcAvailable === null ? 'Checking...' : grpcAvailable ? '✓ Available' : '✗ Unavailable'}
        </h2>
        {grpcAvailable === false && (
          <p className="text-red-700">Make sure the gRPC server is running on port 9001 and the API server is connected.</p>
        )}
      </div>

      {/* Global Error Display */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-800 font-semibold">{error}</p>
        </div>
      )}

      {/* UC-014: Get Manga via gRPC */}
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4 text-blue-600">UC-014: Get Manga via gRPC</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Manga ID:</label>
            <input
              type="text"
              value={getMangaId}
              onChange={(e) => setGetMangaId(e.target.value)}
              className="border rounded px-3 py-2 w-full max-w-xs"
              placeholder="e.g., one-piece, naruto, bleach"
            />
            <p className="text-xs text-gray-500 mt-1">
              Try: one-piece, naruto, attack-on-titan, demon-slayer, death-note
            </p>
          </div>
          <button
            onClick={handleGetManga}
            disabled={loading || !grpcAvailable}
            className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600 disabled:bg-gray-400"
          >
            {loading ? 'Loading...' : 'Get Manga'}
          </button>
          
          {mangaResult && (
            <div className="mt-4 p-4 bg-gray-50 rounded-lg">
              <h3 className="font-semibold mb-2">Result:</h3>
              <pre className="text-sm overflow-x-auto">{JSON.stringify(mangaResult, null, 2)}</pre>
            </div>
          )}
        </div>
      </div>

      {/* UC-015: Search Manga via gRPC */}
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4 text-green-600">UC-015: Search Manga via gRPC</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Search Query:</label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="border rounded px-3 py-2 w-full max-w-md"
              placeholder="Enter search query (e.g., 'One Piece')"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Limit:</label>
            <input
              type="number"
              value={searchLimit}
              onChange={(e) => setSearchLimit(parseInt(e.target.value) || 10)}
              className="border rounded px-3 py-2 w-32"
              min="1"
              max="100"
            />
          </div>
          <button
            onClick={handleSearchManga}
            disabled={loading || !grpcAvailable}
            className="bg-green-500 text-white px-6 py-2 rounded hover:bg-green-600 disabled:bg-gray-400"
          >
            {loading ? 'Searching...' : 'Search Manga'}
          </button>
          
          {searchResults && (
            <div className="mt-4 p-4 bg-gray-50 rounded-lg">
              <h3 className="font-semibold mb-2">Results: {searchResults.total} found</h3>
              <pre className="text-sm overflow-x-auto max-h-96">{JSON.stringify(searchResults, null, 2)}</pre>
            </div>
          )}
        </div>
      </div>

      {/* UC-016: Update Progress via gRPC */}
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4 text-purple-600">UC-016: Update Progress via gRPC</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Manga ID:</label>
            <input
              type="text"
              value={progressMangaId}
              onChange={(e) => setProgressMangaId(e.target.value)}
              className="border rounded px-3 py-2 w-full max-w-xs"
              placeholder="e.g., one-piece, naruto"
            />
            <p className="text-xs text-gray-500 mt-1">
              Must be added to your library first! Go to Browse page to add manga.
            </p>
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Current Chapter:</label>
            <input
              type="number"
              value={progressChapter}
              onChange={(e) => setProgressChapter(parseInt(e.target.value) || 0)}
              className="border rounded px-3 py-2 w-32"
              min="0"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Status:</label>
            <select
              value={progressStatus}
              onChange={(e) => setProgressStatus(e.target.value)}
              className="border rounded px-3 py-2 w-full max-w-xs"
            >
              <option value="reading">Reading</option>
              <option value="completed">Completed</option>
              <option value="plan_to_read">Plan to Read</option>
              <option value="dropped">Dropped</option>
              <option value="on_hold">On Hold</option>
              <option value="re_reading">Re-reading</option>
            </select>
          </div>
          <button
            onClick={handleUpdateProgress}
            disabled={loading || !grpcAvailable}
            className="bg-purple-500 text-white px-6 py-2 rounded hover:bg-purple-600 disabled:bg-gray-400"
          >
            {loading ? 'Updating...' : 'Update Progress'}
          </button>
          
          {progressResult && (
            <div className="mt-4 p-4 bg-gray-50 rounded-lg">
              <h3 className="font-semibold mb-2">Result:</h3>
              <pre className="text-sm overflow-x-auto">{JSON.stringify(progressResult, null, 2)}</pre>
              <p className="mt-2 text-green-600 font-semibold">
                ✓ Progress updated and broadcast via TCP for real-time sync!
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Test Information */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h3 className="text-xl font-semibold mb-3">Test Case Implementation</h3>
        <ul className="list-disc list-inside space-y-2 text-sm">
          <li><strong>UC-014:</strong> Retrieve Manga via gRPC - Fetches manga by ID through gRPC interface</li>
          <li><strong>UC-015:</strong> Search Manga via gRPC - Searches manga with query and pagination</li>
          <li><strong>UC-016:</strong> Update Progress via gRPC - Updates reading progress and triggers TCP broadcast</li>
        </ul>
        <div className="mt-4 p-3 bg-yellow-50 rounded">
          <p className="text-sm"><strong>Note:</strong> Make sure all servers are running:</p>
          <ul className="list-disc list-inside text-sm mt-2">
            <li>TCP Server (port 9000)</li>
            <li>gRPC Server (port 9001)</li>
            <li>API Server (port 8080)</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default GRPCTestPage;
