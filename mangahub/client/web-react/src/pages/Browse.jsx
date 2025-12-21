import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Book } from 'lucide-react';
import mangaService from '../services/mangaService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';

const Browse = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [manga, setManga] = useState([]);
  const [selectedGenres, setSelectedGenres] = useState([]);
  const [selectedStatus, setSelectedStatus] = useState('');
  const [sortBy, setSortBy] = useState('popular');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState(searchParams.get('search') || '');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [pagination, setPagination] = useState(null);
  const itemsPerPage = 25;

  useEffect(() => {
    const searchParam = searchParams.get('search');
    const pageParam = parseInt(searchParams.get('page')) || 1;
    const sortParam = searchParams.get('sort');
    
    // Set sort based on URL parameter, default to 'title' if not specified
    const newSort = sortParam && ['relevant', 'title', 'newest'].includes(sortParam) ? sortParam : 'title';
    
    // Update state synchronously
    setCurrentPage(pageParam);
    setSortBy(newSort);
    
    if (searchParam) {
      // If there's a search parameter, always do search
      setSearchQuery(searchParam);
      handleSearch(searchParam, pageParam, newSort);
    } else {
      // No search parameter, clear search and browse normally
      setSearchQuery('');
      fetchData(pageParam, newSort);
    }
  }, [searchParams.toString()]); // Use toString() to avoid infinite loops

  const fetchData = async (page = 1, currentSort = 'title') => {
    try {
      setLoading(true);
      setError(null);
      
      // Fetch from local database instead of MAL
      const offset = (page - 1) * itemsPerPage;
      const result = await mangaService.searchLocal('', itemsPerPage, offset, currentSort);
      
      setManga(result.manga || []);
      
      // Calculate pagination based on count
      const count = result.count || 0;
      setTotalItems(count);
      setTotalPages(Math.ceil(count / itemsPerPage) || 1);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError(err.message);
      setManga([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async (query, page = 1, currentSort = 'relevant') => {
    if (!query.trim()) {
      fetchData(page, currentSort);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      // Search in local database
      const offset = (page - 1) * itemsPerPage;
      const result = await mangaService.searchLocal(query, itemsPerPage, offset, currentSort);
      
      setManga(result.manga || []);
      
      // Calculate pagination based on count
      const count = result.count || 0;
      setTotalItems(count);
      setTotalPages(Math.ceil(count / itemsPerPage) || 1);
    } catch (err) {
      console.error('Search error:', err);
      setError(err.message);
      setManga([]);
      setTotalItems(0);
      setTotalPages(1);
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (page) => {
    if (page < 1 || page > totalPages) return;
    
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('page', page.toString());
    setSearchParams(newSearchParams);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const filteredManga = manga.filter(m => {
    // Genre filter - manga must have ALL selected genres
    if (selectedGenres.length > 0) {
      if (!m.genres) return false;
      const hasAllGenres = selectedGenres.every(genre => m.genres.includes(genre));
      if (!hasAllGenres) return false;
    }
    
    // Status filter
    if (selectedStatus) {
      if (!m.status) return false;
      if (selectedStatus === 'ongoing' && m.status !== 'ongoing' && m.status !== 'publishing') return false;
      if (selectedStatus === 'completed' && m.status !== 'completed') return false;
    }
    
    return true;
  });

  // No need for client-side sorting anymore - backend handles it
  const displayedManga = filteredManga;

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark py-8 transition-colors">
      <div className="container mx-auto px-4">
        <div className="flex flex-col lg:flex-row gap-6">
          {/* Sidebar */}
          <div className="w-full lg:w-64 flex-shrink-0">
            <div className="sticky top-24 bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 p-6 space-y-6">
              {/* Sort By */}
              <div>
                <label className="block text-sm font-bold text-zinc-900 dark:text-white mb-3">Sort By</label>
                <div className="space-y-1">
                  {[
                    { value: 'title', label: 'Title (A-Z)' },
                    { value: 'newest', label: 'Newest' }
                  ].map(({ value, label }) => (
                    <button
                      key={value}
                      onClick={() => {
                        const newSearchParams = new URLSearchParams(searchParams);
                        newSearchParams.set('sort', value);
                        newSearchParams.set('page', '1'); // Reset to page 1 when changing sort
                        setSearchParams(newSearchParams);
                      }}
                      className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                        sortBy === value
                          ? 'bg-primary/10 text-primary font-semibold'
                          : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                      }`}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Genres */}
              <div>
                <label className="block text-sm font-bold text-zinc-900 dark:text-white mb-3">
                  Genres {selectedGenres.length > 0 && <span className="text-xs text-primary">({selectedGenres.length})</span>}
                </label>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  <button
                    onClick={() => setSelectedGenres([])}
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                      selectedGenres.length === 0
                        ? 'bg-primary/10 text-primary font-semibold'
                        : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                    }`}
                  >
                    All Genres
                  </button>
                  {['Action', 'Adventure', 'Award Winning', 'Comedy', 'Drama', 'Fantasy', 'Horror', 'Mystery', 'Psychological', 'Romance', 'Sci-Fi', 'Slice of Life', 'Sports', 'Supernatural', 'Thriller'].map(genre => (
                    <label key={genre} className="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800 px-3 py-2 rounded-lg transition-colors">
                      <input
                        type="checkbox"
                        checked={selectedGenres.includes(genre)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setSelectedGenres([...selectedGenres, genre]);
                          } else {
                            setSelectedGenres(selectedGenres.filter(g => g !== genre));
                          }
                        }}
                        className="w-4 h-4 rounded border-zinc-300 dark:border-zinc-600 text-primary focus:ring-primary"
                      />
                      <span className="text-sm text-zinc-600 dark:text-zinc-400">{genre}</span>
                    </label>
                  ))}
                </div>
              </div>

              {/* Status */}
              <div>
                <label className="block text-sm font-bold text-zinc-900 dark:text-white mb-3">Status</label>
                <div className="space-y-1">
                  {[
                    { value: '', label: 'All' },
                    { value: 'ongoing', label: 'Ongoing' },
                    { value: 'completed', label: 'Completed' },
                  ].map(({ value, label }) => (
                    <button
                      key={value}
                      onClick={() => setSelectedStatus(value)}
                      className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                        selectedStatus === value
                          ? 'bg-primary/10 text-primary font-semibold'
                          : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                      }`}
                    >
                      {label}
                    </button>
                  ))}
                </div>
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
              <h1 className="text-4xl font-black text-zinc-900 dark:text-white mb-2 tracking-tight">Browse</h1>
              <p className="text-zinc-600 dark:text-zinc-400">
                Showing <span className="font-semibold text-zinc-900 dark:text-white">{displayedManga.length > 0 ? (currentPage - 1) * itemsPerPage + 1 : 0}-{Math.min(currentPage * itemsPerPage, totalItems)}</span> of <span className="font-semibold text-zinc-900 dark:text-white">{totalItems.toLocaleString()}</span> manga
                {selectedGenres.length > 0 && <span> in <span className="font-semibold text-primary">{selectedGenres.join(', ')}</span></span>}
                {selectedStatus && <span> • <span className="font-semibold text-primary">{selectedStatus === 'ongoing' ? 'Ongoing' : selectedStatus === 'completed' ? 'Completed' : 'On Hiatus'}</span></span>}
              </p>
            </motion.div>

            {/* Manga Grid */}
            {loading ? (
              <LoadingSpinner message="Loading manga..." />
            ) : error ? (
              <div className="text-center py-20 bg-red-50 dark:bg-red-900/10 rounded-2xl border border-red-100 dark:border-red-900/20">
                <p className="text-red-600 dark:text-red-400 mb-4 text-lg font-medium">{error}</p>
                <button
                  onClick={fetchData}
                  className="px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition font-semibold"
                >
                  Try Again
                </button>
              </div>
            ) : displayedManga.length > 0 ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.2 }}
                className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-5 gap-4"
              >
                {displayedManga.map((m) => (
                  <MangaCard key={m.id} manga={m} />
                ))}
              </motion.div>
            ) : (
              <div className="text-center py-20 bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800">
                <Book className="w-16 h-16 mx-auto mb-4 text-zinc-300 dark:text-zinc-700" />
                <h3 className="text-xl font-bold text-zinc-900 dark:text-white mb-2">No manga found</h3>
                <p className="text-zinc-500 dark:text-zinc-400 mb-6">We couldn't find any manga matching your filters.</p>
                <button
                  onClick={() => {
                    setSelectedGenres([]);
                    setSelectedStatus('');
                    setSearchQuery('');
                    setSortBy('popular');
                  }}
                  className="px-6 py-2.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition shadow-lg shadow-primary/25"
                >
                  Reset Filters
                </button>
              </div>
            )}

            {/* Pagination */}
            {!loading && displayedManga.length > 0 && (
              <div className="mt-12 flex items-center justify-center gap-2 flex-wrap">
                <button 
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className="px-4 py-2 bg-white dark:bg-[#191022] border border-zinc-200 dark:border-zinc-800 rounded-lg text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                
                {/* Page Numbers */}
                {(() => {
                  const pages = [];
                  const maxVisible = 5;
                  let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
                  let endPage = Math.min(totalPages, startPage + maxVisible - 1);
                  
                  if (endPage - startPage < maxVisible - 1) {
                    startPage = Math.max(1, endPage - maxVisible + 1);
                  }
                  
                  if (startPage > 1) {
                    pages.push(
                      <button
                        key={1}
                        onClick={() => handlePageChange(1)}
                        className="px-4 py-2 rounded-lg transition bg-white dark:bg-[#191022] border border-zinc-200 dark:border-zinc-800 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800"
                      >
                        1
                      </button>
                    );
                    if (startPage > 2) {
                      pages.push(<span key="ellipsis1" className="px-2 text-zinc-400">...</span>);
                    }
                  }
                  
                  for (let i = startPage; i <= endPage; i++) {
                    pages.push(
                      <button
                        key={i}
                        onClick={() => handlePageChange(i)}
                        className={`px-4 py-2 rounded-lg transition ${
                          i === currentPage
                            ? 'bg-primary text-white'
                            : 'bg-white dark:bg-[#191022] border border-zinc-200 dark:border-zinc-800 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800'
                        }`}
                      >
                        {i}
                      </button>
                    );
                  }
                  
                  if (endPage < totalPages) {
                    if (endPage < totalPages - 1) {
                      pages.push(<span key="ellipsis2" className="px-2 text-zinc-400">...</span>);
                    }
                    pages.push(
                      <button
                        key={totalPages}
                        onClick={() => handlePageChange(totalPages)}
                        className="px-4 py-2 rounded-lg transition bg-white dark:bg-[#191022] border border-zinc-200 dark:border-zinc-800 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800"
                      >
                        {totalPages}
                      </button>
                    );
                  }
                  
                  return pages;
                })()}
                
                <button 
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage === totalPages}
                  className="px-4 py-2 bg-white dark:bg-[#191022] border border-zinc-200 dark:border-zinc-800 rounded-lg text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Browse;
