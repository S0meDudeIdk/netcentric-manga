import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Book, TrendingUp, Search, Library, ArrowRight, Sparkles } from 'lucide-react';
import mangaService from '../services/mangaService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';
import authService from '../services/authService';

const Home = () => {
  const [popularManga, setPopularManga] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const isAuthenticated = authService.isAuthenticated();

  useEffect(() => {
    const fetchPopularManga = async () => {
      try {
        setLoading(true);
        // Fetch from MyAnimeList top manga
        const data = await mangaService.getTopMAL(1, 6);
        setPopularManga(data.data || []);
      } catch (err) {
        console.error('Error fetching popular manga:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchPopularManga();
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50">
      {/* Hero Section */}
      <section className="container mx-auto px-4 py-16">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="text-center"
        >
          <div className="flex items-center justify-center gap-3 mb-4">
            <Book className="w-16 h-16 text-blue-600" />
            <h1 className="text-5xl md:text-6xl font-bold text-gray-900">
              MangaHub
            </h1>
          </div>
          
          <p className="text-xl md:text-2xl text-gray-600 mb-8 max-w-2xl mx-auto">
            Discover, browse, and read manga freely. Create an account to track your progress and build your personal library!
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              to="/browse"
              className="inline-flex items-center gap-2 px-8 py-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition font-semibold text-lg shadow-lg hover:shadow-xl"
            >
              <Book className="w-6 h-6" />
              <span>Browse Manga</span>
              <ArrowRight className="w-5 h-5" />
            </Link>
            
            <Link
              to="/search"
              className="inline-flex items-center gap-2 px-8 py-4 bg-white text-blue-600 border-2 border-blue-600 rounded-lg hover:bg-blue-50 transition font-semibold text-lg shadow-lg hover:shadow-xl"
            >
              <Search className="w-6 h-6" />
              <span>Search</span>
            </Link>
          </div>
        </motion.div>
      </section>

      {/* Features Section */}
      <section className="container mx-auto px-4 py-12">
        <h2 className="text-3xl font-bold text-center text-gray-900 mb-8">
          Browse Freely, Track with an Account
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="bg-white rounded-lg p-6 shadow-md hover:shadow-xl transition"
          >
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
              <Book className="w-6 h-6 text-green-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 mb-2">Free Browsing</h3>
            <p className="text-gray-600">
              Browse and read any manga without creating an account. No login required!
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="bg-white rounded-lg p-6 shadow-md hover:shadow-xl transition"
          >
            <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
              <TrendingUp className="w-6 h-6 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 mb-2">Track Progress</h3>
            <p className="text-gray-600">
              Sign up to save your reading progress, pick up where you left off, and track chapters read.
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            className="bg-white rounded-lg p-6 shadow-md hover:shadow-xl transition"
          >
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
              <Library className="w-6 h-6 text-purple-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 mb-2">Personal Library</h3>
            <p className="text-gray-600">
              Create an account to build your collection, organize with playlists, and manage your manga.
            </p>
          </motion.div>
        </div>
      </section>

      {/* Remove old third feature and update the section below */}
      <section className="container mx-auto px-4 py-12">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg p-8 shadow-md"
          >
            <h3 className="text-2xl font-bold text-gray-900 mb-4">For Everyone</h3>
            <ul className="space-y-3">
              <li className="flex items-start gap-2">
                <span className="text-green-600 mt-1">✓</span>
                <span className="text-gray-700">Browse entire manga collection</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-green-600 mt-1">✓</span>
                <span className="text-gray-700">Search by title, author, or genre</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-green-600 mt-1">✓</span>
                <span className="text-gray-700">View full manga details and descriptions</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-green-600 mt-1">✓</span>
                <span className="text-gray-700">Filter and sort manga</span>
              </li>
            </ul>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.5 }}
            className="bg-gradient-to-br from-purple-50 to-purple-100 rounded-lg p-8 shadow-md"
          >
            <h3 className="text-2xl font-bold text-gray-900 mb-4">With an Account</h3>
            <ul className="space-y-3">
              <li className="flex items-start gap-2">
                <span className="text-purple-600 mt-1">★</span>
                <span className="text-gray-700">Save manga to your personal library</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-600 mt-1">★</span>
                <span className="text-gray-700">Track reading progress and chapters</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-600 mt-1">★</span>
                <span className="text-gray-700">Continue where you left off</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-600 mt-1">★</span>
                <span className="text-gray-700">Create reading lists and playlists</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-600 mt-1">★</span>
                <span className="text-gray-700">Get personalized recommendations</span>
              </li>
            </ul>
          </motion.div>
        </div>
      </section>

      {/* Popular Manga Section */}
      <section className="container mx-auto px-4 py-12">
        <div className="flex items-center justify-between mb-8">
          <h2 className="text-3xl font-bold text-gray-900">Popular Manga</h2>
          <Link
            to="/browse"
            className="text-blue-600 hover:text-blue-700 font-semibold flex items-center gap-2"
          >
            View All
            <ArrowRight className="w-5 h-5" />
          </Link>
        </div>

        {loading ? (
          <LoadingSpinner message="Loading popular manga..." />
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-600 mb-4">{error}</p>
            <button
              onClick={() => window.location.reload()}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Try Again
            </button>
          </div>
        ) : popularManga.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-6">
            {popularManga.map((manga) => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </div>
        ) : (
          <div className="text-center py-12 text-gray-600">
            <Book className="w-16 h-16 mx-auto mb-4 text-gray-400" />
            <p>No manga found. Add some manga to get started!</p>
          </div>
        )}
      </section>

      {/* CTA Section */}
      {!isAuthenticated && (
        <section className="container mx-auto px-4 py-16">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6 }}
            className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-2xl p-12 text-center text-white shadow-2xl"
          >
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Want to track your reading progress?
            </h2>
            <p className="text-xl mb-8 opacity-90">
              Create a free account to save your progress, build your library, and never lose your place!
            </p>
            <Link
              to="/register"
              className="inline-flex items-center gap-2 px-8 py-4 bg-white text-blue-600 rounded-lg hover:bg-gray-100 transition font-semibold text-lg shadow-lg"
            >
              <span>Create Free Account</span>
              <ArrowRight className="w-5 h-5" />
            </Link>
          </motion.div>
        </section>
      )}
    </div>
  );
};

export default Home;
