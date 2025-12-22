import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Library as LibraryIcon, Book, Filter, CheckCircle2, Clock, XCircle, PauseCircle, Search, ChevronDown, Plus, RotateCw } from 'lucide-react';
import libraryService from '../services/libraryService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';

const Library = () => {
  const [library, setLibrary] = useState([]);
  const [selectedStatus, setSelectedStatus] = useState('all');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchLibraryData();
  }, []);

  const fetchLibraryData = async () => {
    try {
      setLoading(true);
      const libraryData = await libraryService.getLibrary();

      // Flatten the library data from different status arrays into one
      const flatLibrary = [
        ...(libraryData.reading || []),
        ...(libraryData.completed || []),
        ...(libraryData.plan_to_read || []),
        ...(libraryData.dropped || []),
        ...(libraryData.on_hold || []),
        ...(libraryData.re_reading || [])
      ].map(item => ({
        ...item,
        // Transform to match expected structure - manga details are now directly on the progress object
        manga: {
          id: item.manga_id,
          title: item.title,
          author: item.author,
          cover_url: item.cover_url
        }
      }));

      setLibrary(flatLibrary);
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
    re_reading: library.filter(item => item.status === 're_reading').length,
  };

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark py-8 transition-colors">
      <div className="container mx-auto px-4">
        <div className="flex flex-col lg:flex-row gap-6">
          {/* Sidebar */}
          <div className="w-full lg:w-64 flex-shrink-0">
            <div className="sticky top-24 bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 p-6">
              {/* User Profile */}
              <div className="flex items-center gap-3 mb-6 pb-6 border-b border-zinc-200 dark:border-zinc-800">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
                  <LibraryIcon className="w-6 h-6 text-primary" />
                </div>
                <div>
                  <p className="font-bold text-zinc-900 dark:text-white">MangaHub User</p>
                  <p className="text-sm text-zinc-500 dark:text-zinc-400">My Library</p>
                </div>
              </div>

              {/* Status Navigation */}
              <div className="space-y-1">
                {[
                  { value: 'all', label: 'All', icon: Filter, count: statusCounts.all },
                  { value: 'reading', label: 'Reading', icon: Book, count: statusCounts.reading },
                  { value: 'completed', label: 'Completed', icon: CheckCircle2, count: statusCounts.completed },
                  { value: 'plan_to_read', label: 'Plan to Read', icon: Clock, count: statusCounts.plan_to_read },
                  { value: 'on_hold', label: 'On Hold', icon: PauseCircle, count: statusCounts.on_hold },
                  { value: 'dropped', label: 'Dropped', icon: XCircle, count: statusCounts.dropped },
                  { value: 're_reading', label: 'Re-Reading', icon: RotateCw, count: statusCounts.re_reading },
                ].map(({ value, label, icon: Icon, count }) => (
                  <button
                    key={value}
                    onClick={() => setSelectedStatus(value)}
                    className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm transition-colors ${
                      selectedStatus === value
                        ? 'bg-primary text-white'
                        : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <Icon className="w-4 h-4" />
                      <span className="font-medium">{label}</span>
                    </div>
                    <span className={`px-2 py-0.5 rounded-md text-xs font-semibold ${
                      selectedStatus === value 
                        ? 'bg-white/20' 
                        : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400'
                    }`}>
                      {count}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Main Content */}
          <div className="flex-1">
            {/* Header */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mb-8"
            >
              <h1 className="text-4xl font-black text-zinc-900 dark:text-white mb-2 tracking-tight">
                {selectedStatus === 'all' ? 'Reading' : selectedStatus.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())} ({filteredLibrary.length})
              </h1>
              <p className="text-zinc-600 dark:text-zinc-400">
                {selectedStatus === 'all' 
                  ? 'All the manga you are currently reading.' 
                  : `All the manga in ${selectedStatus.replace('_', ' ')}.`}
              </p>
            </motion.div>

            {/* Search and Sort */}
            <div className="flex gap-4 mb-6">
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
                <input
                  type="text"
                  placeholder="Search in your library..."
                  className="w-full pl-10 pr-4 py-2.5 bg-white dark:bg-[#211c27] border border-zinc-200 dark:border-zinc-700 rounded-lg focus:ring-2 focus:ring-primary/50 focus:border-primary text-zinc-900 dark:text-white placeholder:text-zinc-400 outline-none transition-all text-sm"
                />
              </div>
              <div className="flex gap-2">
                <button className="px-4 py-2.5 bg-white dark:bg-[#211c27] border border-zinc-200 dark:border-zinc-700 rounded-lg text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition text-sm font-medium">
                  Genre
                  <ChevronDown className="w-4 h-4 inline ml-1" />
                </button>
                <button className="px-4 py-2.5 bg-white dark:bg-[#211c27] border border-zinc-200 dark:border-zinc-700 rounded-lg text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition text-sm font-medium">
                  Sort By
                  <ChevronDown className="w-4 h-4 inline ml-1" />
                </button>
              </div>
            </div>

            {loading ? (
              <LoadingSpinner message="Loading your library..." />
            ) : error ? (
              <div className="text-center py-20 bg-red-50 dark:bg-red-900/10 rounded-2xl border border-red-100 dark:border-red-900/20">
                <p className="text-red-600 dark:text-red-400 mb-4 text-lg font-medium">{error}</p>
                <button
                  onClick={fetchLibraryData}
                  className="px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition font-semibold"
                >
                  Try Again
                </button>
              </div>
            ) : filteredLibrary.length > 0 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.1 }}
                className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-5 gap-4"
              >
                {filteredLibrary.map((item, idx) => (
                  <MangaCard 
                    key={item.manga_id} 
                    manga={item.manga} 
                    index={idx}
                    currentChapter={item.current_chapter}
                  />
                ))}
              </motion.div>
            ) : (
              <div className="text-center py-20 bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800">
                <Book className="w-16 h-16 mx-auto mb-4 text-zinc-300 dark:text-zinc-700" />
                <h3 className="text-xl font-bold text-zinc-900 dark:text-white mb-2">
                  {selectedStatus === 'all' ? 'Your library is empty' : `No manga in ${selectedStatus.replace('_', ' ')}`}
                </h3>
                <p className="text-zinc-500 dark:text-zinc-400 max-w-md mx-auto mb-6">
                  {selectedStatus === 'all'
                    ? 'Start adding manga to your library to track your reading progress.'
                    : 'Try selecting a different status or add some manga to your library.'
                  }
                </p>
                <Link
                  to="/browse"
                  className="inline-flex items-center gap-2 px-6 py-2.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition shadow-lg shadow-primary/25"
                >
                  <Book className="w-4 h-4" />
                  <span>Browse Manga</span>
                </Link>
              </div>
            )}

            {/* Add New Manga Button */}
            {filteredLibrary.length > 0 && (
              <div className="mt-8">
                <Link
                  to="/browse"
                  className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-white dark:bg-[#211c27] border-2 border-dashed border-zinc-300 dark:border-zinc-700 rounded-xl text-zinc-600 dark:text-zinc-400 hover:border-primary hover:text-primary transition font-semibold"
                >
                  <Plus className="w-5 h-5" />
                  <span>Add New Manga</span>
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Library;
