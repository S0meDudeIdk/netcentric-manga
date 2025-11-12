import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Library as LibraryIcon, TrendingUp, BarChart3, Book, Sparkles } from 'lucide-react';
import userService from '../services/userService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';

const Library = () => {
  const [library, setLibrary] = useState([]);
  const [stats, setStats] = useState(null);
  const [recommendations, setRecommendations] = useState([]);
  const [selectedStatus, setSelectedStatus] = useState('all');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchLibraryData();
  }, []);

  const fetchLibraryData = async () => {
    try {
      setLoading(true);
      const [libraryData, statsData, recsData] = await Promise.all([
        userService.getLibrary(),
        userService.getLibraryStats().catch(() => null),
        userService.getRecommendations(5).catch(() => ({ manga: [] }))
      ]);
      
      setLibrary(libraryData.library || []);
      setStats(statsData);
      setRecommendations(recsData.manga || []);
    } catch (err) {
      console.error('Error fetching library:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filteredLibrary = library.filter(item => {
    if (selectedStatus === 'all') return true;
    return item.status === selectedStatus;
  });

  const statusCounts = {
    all: library.length,
    reading: library.filter(item => item.status === 'reading').length,
    completed: library.filter(item => item.status === 'completed').length,
    plan_to_read: library.filter(item => item.status === 'plan_to_read').length,
    on_hold: library.filter(item => item.status === 'on_hold').length,
    dropped: library.filter(item => item.status === 'dropped').length,
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8"
        >
          <div className="flex items-center gap-3 mb-4">
            <LibraryIcon className="w-8 h-8 text-blue-600" />
            <h1 className="text-4xl font-bold text-gray-900">My Library</h1>
          </div>
          <p className="text-gray-600">Track your manga reading progress</p>
        </motion.div>

        {loading ? (
          <LoadingSpinner message="Loading your library..." />
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-600 mb-4">{error}</p>
            <button
              onClick={fetchLibraryData}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Try Again
            </button>
          </div>
        ) : (
          <>
            {/* Stats Cards */}
            {stats && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8"
              >
                <div className="bg-white rounded-lg shadow-md p-6">
                  <div className="flex items-center justify-between mb-2">
                    <Book className="w-8 h-8 text-blue-600" />
                    <span className="text-3xl font-bold text-gray-900">{stats.total_manga || 0}</span>
                  </div>
                  <p className="text-gray-600">Total Manga</p>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                  <div className="flex items-center justify-between mb-2">
                    <TrendingUp className="w-8 h-8 text-green-600" />
                    <span className="text-3xl font-bold text-gray-900">{stats.total_chapters_read || 0}</span>
                  </div>
                  <p className="text-gray-600">Chapters Read</p>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                  <div className="flex items-center justify-between mb-2">
                    <BarChart3 className="w-8 h-8 text-purple-600" />
                    <span className="text-3xl font-bold text-gray-900">{statusCounts.reading}</span>
                  </div>
                  <p className="text-gray-600">Currently Reading</p>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                  <div className="flex items-center justify-between mb-2">
                    <Sparkles className="w-8 h-8 text-yellow-600" />
                    <span className="text-3xl font-bold text-gray-900">{statusCounts.completed}</span>
                  </div>
                  <p className="text-gray-600">Completed</p>
                </div>
              </motion.div>
            )}

            {/* Status Filter */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="bg-white rounded-lg shadow-md p-6 mb-8"
            >
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Filter by Status</h2>
              <div className="flex flex-wrap gap-2">
                {[
                  { value: 'all', label: 'All', color: 'gray' },
                  { value: 'reading', label: 'Reading', color: 'blue' },
                  { value: 'completed', label: 'Completed', color: 'green' },
                  { value: 'plan_to_read', label: 'Plan to Read', color: 'yellow' },
                  { value: 'on_hold', label: 'On Hold', color: 'orange' },
                  { value: 'dropped', label: 'Dropped', color: 'red' },
                ].map(({ value, label, color }) => (
                  <button
                    key={value}
                    onClick={() => setSelectedStatus(value)}
                    className={`px-4 py-2 rounded-lg font-semibold transition ${
                      selectedStatus === value
                        ? `bg-${color}-600 text-white`
                        : `bg-${color}-100 text-${color}-700 hover:bg-${color}-200`
                    }`}
                    style={
                      selectedStatus === value
                        ? { backgroundColor: `var(--color-${color}-600, #3b82f6)` }
                        : {}
                    }
                  >
                    {label} ({statusCounts[value]})
                  </button>
                ))}
              </div>
            </motion.div>

            {/* Library Grid */}
            {filteredLibrary.length > 0 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.3 }}
              >
                <h2 className="text-2xl font-bold text-gray-900 mb-6">
                  {selectedStatus === 'all' ? 'All Manga' : `${selectedStatus.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}`}
                  {' '}({filteredLibrary.length})
                </h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
                  {filteredLibrary.map((item) => (
                    <MangaCard key={item.manga_id} manga={item.manga} />
                  ))}
                </div>
              </motion.div>
            ) : (
              <div className="text-center py-12">
                <Book className="w-16 h-16 mx-auto mb-4 text-gray-400" />
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  {selectedStatus === 'all' ? 'Your library is empty' : `No manga in ${selectedStatus.replace('_', ' ')}`}
                </h3>
                <p className="text-gray-600">
                  {selectedStatus === 'all' 
                    ? 'Start adding manga to your library to track your reading progress'
                    : 'Try selecting a different status filter'
                  }
                </p>
              </div>
            )}

            {/* Recommendations */}
            {recommendations.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4 }}
                className="mt-12"
              >
                <div className="flex items-center gap-2 mb-6">
                  <Sparkles className="w-6 h-6 text-yellow-600" />
                  <h2 className="text-2xl font-bold text-gray-900">Recommended for You</h2>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
                  {recommendations.map((manga) => (
                    <MangaCard key={manga.id} manga={manga} />
                  ))}
                </div>
              </motion.div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default Library;
