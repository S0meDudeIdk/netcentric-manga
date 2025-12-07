import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Search as SearchIcon, Book, X } from 'lucide-react';
import mangaService from '../services/mangaService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';

const Search = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = async (e) => {
    e.preventDefault();
    
    if (!query.trim()) {
      return;
    }

    setLoading(true);
    setError(null);
    setHasSearched(true);

    try {
      const data = await mangaService.searchMAL(query);
      setResults(data.data || []);
    } catch (err) {
      console.error('Search error:', err);
      setError(err.message);
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    setQuery('');
    setResults([]);
    setHasSearched(false);
    setError(null);
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
            <SearchIcon className="w-8 h-8 text-blue-600" />
            <h1 className="text-4xl font-bold text-gray-900">Search Manga</h1>
          </div>
          <p className="text-gray-600">
            Search by title, author, or genre from MyAnimeList
          </p>
        </motion.div>

        {/* Search Bar */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="bg-white rounded-lg shadow-md p-6 mb-8"
        >
          <form onSubmit={handleSearch} className="flex gap-4">
            <div className="flex-1 relative">
              <SearchIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search for manga..."
                className="w-full pl-12 pr-12 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-lg"
              />
              {query && (
                <button
                  type="button"
                  onClick={handleClear}
                  className="absolute right-4 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  <X className="w-5 h-5" />
                </button>
              )}
            </div>
            <button
              type="submit"
              disabled={!query.trim() || loading}
              className="px-8 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition font-semibold disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <SearchIcon className="w-5 h-5" />
              <span>Search</span>
            </button>
          </form>

          {/* Quick Suggestions */}
          {!hasSearched && (
            <div className="mt-4">
              <p className="text-sm text-gray-600 mb-2">Try searching for:</p>
              <div className="flex flex-wrap gap-2">
                {['One Piece', 'Naruto', 'Attack on Titan', 'Death Note', 'Dragon Ball'].map((suggestion) => (
                  <button
                    key={suggestion}
                    onClick={() => {
                      setQuery(suggestion);
                      // Trigger search automatically
                      setTimeout(() => {
                        const event = { preventDefault: () => {} };
                        setQuery(suggestion);
                        handleSearch(event);
                      }, 100);
                    }}
                    className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm hover:bg-gray-200 transition"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>
          )}
        </motion.div>

        {/* Results */}
        {loading ? (
          <LoadingSpinner message="Searching..." />
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-600 mb-4">{error}</p>
            <button
              onClick={handleSearch}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Try Again
            </button>
          </div>
        ) : hasSearched ? (
          results.length > 0 ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.2 }}
            >
              <div className="mb-6">
                <h2 className="text-2xl font-bold text-gray-900">
                  Search Results ({results.length})
                </h2>
                <p className="text-gray-600">Found {results.length} manga matching "{query}"</p>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
                {results.map((manga) => (
                  <MangaCard key={manga.id} manga={manga} />
                ))}
              </div>
            </motion.div>
          ) : (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-12"
            >
              <Book className="w-16 h-16 mx-auto mb-4 text-gray-400" />
              <h3 className="text-xl font-semibold text-gray-900 mb-2">No manga found</h3>
              <p className="text-gray-600 mb-4">
                We couldn't find any manga matching "{query}"
              </p>
              <p className="text-sm text-gray-500">
                Try different keywords or check your spelling
              </p>
              <button
                onClick={handleClear}
                className="mt-6 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Clear Search
              </button>
            </motion.div>
          )
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-16"
          >
            <SearchIcon className="w-20 h-20 mx-auto mb-4 text-gray-300" />
            <h3 className="text-2xl font-semibold text-gray-700 mb-2">
              Start searching for manga
            </h3>
            <p className="text-gray-500">
              Enter a title, author, or genre in the search box above
            </p>
          </motion.div>
        )}
      </div>
    </div>
  );
};

export default Search;
