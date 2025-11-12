import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Book, Filter, TrendingUp, Globe } from 'lucide-react';
import mangaService from '../services/mangaService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';

const Browse = () => {
  const [manga, setManga] = useState([]);
  const [genres, setGenres] = useState([]);
  const [selectedGenre, setSelectedGenre] = useState('');
  const [sortBy, setSortBy] = useState('popular');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [dataSource, setDataSource] = useState('mal'); // 'local' or 'mal'

  useEffect(() => {
    fetchData();
  }, [dataSource]);

  const fetchData = async () => {
    try {
      setLoading(true);
      if (dataSource === 'mal') {
        // Fetch from MyAnimeList
        const malData = await mangaService.getTopMAL(1, 25);
        setManga(malData.data || []);
        setGenres([]); // MAL data already has genres in each manga
      } else {
        // Fetch from local database
        const [mangaData, genresData] = await Promise.all([
          mangaService.getPopularManga(),
          mangaService.getGenres().catch(() => ({ genres: [] }))
        ]);
        setManga(mangaData.manga || []);
        setGenres(genresData.genres || []);
      }
    } catch (err) {
      console.error('Error fetching data:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filteredManga = manga.filter(m => {
    if (!selectedGenre) return true;
    return m.genres && m.genres.includes(selectedGenre);
  });

  const sortedManga = [...filteredManga].sort((a, b) => {
    switch (sortBy) {
      case 'popular':
        return (b.rating || 0) - (a.rating || 0);
      case 'title':
        return a.title.localeCompare(b.title);
      case 'chapters':
        return (b.total_chapters || 0) - (a.total_chapters || 0);
      case 'year':
        return (b.publication_year || 0) - (a.publication_year || 0);
      default:
        return 0;
    }
  });

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8"
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <Book className="w-8 h-8 text-blue-600" />
              <h1 className="text-4xl font-bold text-gray-900">Browse Manga</h1>
            </div>
            
            {/* Data Source Toggle */}
            <div className="flex items-center gap-2 bg-white rounded-lg shadow-md p-2">
              <button
                onClick={() => setDataSource('mal')}
                className={`flex items-center gap-2 px-4 py-2 rounded-md transition-colors ${
                  dataSource === 'mal'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                <Globe className="w-4 h-4" />
                MyAnimeList
              </button>
              <button
                onClick={() => setDataSource('local')}
                className={`flex items-center gap-2 px-4 py-2 rounded-md transition-colors ${
                  dataSource === 'local'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                <Book className="w-4 h-4" />
                Local
              </button>
            </div>
          </div>
          <p className="text-gray-600">
            {dataSource === 'mal' 
              ? 'Discover top manga from MyAnimeList' 
              : 'Browse our local manga collection'}
          </p>
        </motion.div>

        {/* Filters */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="bg-white rounded-lg shadow-md p-6 mb-8"
        >
          <div className="flex items-center gap-2 mb-4">
            <Filter className="w-5 h-5 text-gray-600" />
            <h2 className="text-lg font-semibold text-gray-900">Filters & Sorting</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Genre Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Genre
              </label>
              <select
                value={selectedGenre}
                onChange={(e) => setSelectedGenre(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Genres</option>
                {genres.map((genre) => (
                  <option key={genre} value={genre}>
                    {genre}
                  </option>
                ))}
              </select>
            </div>

            {/* Sort By */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Sort By
              </label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="popular">Most Popular</option>
                <option value="title">Title (A-Z)</option>
                <option value="chapters">Most Chapters</option>
                <option value="year">Newest</option>
              </select>
            </div>
          </div>

          {/* Results Count */}
          <div className="mt-4 flex items-center gap-2 text-sm text-gray-600">
            <TrendingUp className="w-4 h-4" />
            <span>
              Showing {sortedManga.length} manga
              {selectedGenre && ` in ${selectedGenre}`}
            </span>
          </div>
        </motion.div>

        {/* Manga Grid */}
        {loading ? (
          <LoadingSpinner message="Loading manga..." />
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-600 mb-4">{error}</p>
            <button
              onClick={fetchData}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Try Again
            </button>
          </div>
        ) : sortedManga.length > 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
            className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
          >
            {sortedManga.map((m) => (
              <MangaCard key={m.id} manga={m} />
            ))}
          </motion.div>
        ) : (
          <div className="text-center py-12">
            <Book className="w-16 h-16 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-600 text-lg mb-2">No manga found</p>
            <p className="text-gray-500">Try changing your filters</p>
            <button
              onClick={() => {
                setSelectedGenre('');
                setSortBy('popular');
              }}
              className="mt-4 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Reset Filters
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Browse;
