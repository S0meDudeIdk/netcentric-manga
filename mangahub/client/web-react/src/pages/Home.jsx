import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Book, TrendingUp, Library, ArrowRight } from 'lucide-react';
import mangaService from '../services/mangaService';
import userService from '../services/userService';
import MangaCard from '../components/MangaCard';
import LoadingSpinner from '../components/LoadingSpinner';
import authService from '../services/authService';

const Home = () => {
  const navigate = useNavigate();
  const [trendingManga, setTrendingManga] = useState([]);
  const [recentlyAdded, setRecentlyAdded] = useState([]);
  const [continueReading, setContinueReading] = useState([]);
  const [loading, setLoading] = useState(true);
  const isAuthenticated = authService.isAuthenticated();

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        // Fetch trending (top rated manga)
        const trendingData = await mangaService.getTopMAL(1, 6);
        setTrendingManga(trendingData.data || []);
        
        // Fetch recently added (sorted by start_date descending for newest manga)
        const recentData = await mangaService.getTopMAL(1, 6, 'start_date', 'desc');
        setRecentlyAdded(recentData.data || []);
        
        // Fetch continue reading if authenticated
        if (isAuthenticated) {
          try {
            const libraryData = await userService.getLibrary();
            const readingManga = libraryData.library?.filter(item => item.status === 'reading').slice(0, 3) || [];
            setContinueReading(readingManga);
          } catch (err) {
            // Silently handle library errors (e.g., token expired, user not found)
            // This is expected for logged-out users
            if (err.response?.status !== 401) {
              console.error('Error fetching library:', err);
            }
          }
        }
      } catch (err) {
        console.error('Error fetching manga:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [isAuthenticated]);

  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-background-light dark:bg-background-dark py-12 lg:py-16">
        <div className="absolute inset-0 illustration-bg opacity-50"></div>
        <div className="container mx-auto px-4 relative z-10">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="max-w-4xl mx-auto text-center"
          >
            <h1 className="text-4xl md:text-5xl font-black text-zinc-900 dark:text-white mb-6 tracking-tight leading-tight">
              Welcome back!
            </h1>
          </motion.div>
        </div>
      </section>

      {/* Continue Reading Section */}
      {isAuthenticated && continueReading.length > 0 && (
        <section className="py-12 bg-white dark:bg-[#191022]">
          <div className="container mx-auto px-4">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-black text-zinc-900 dark:text-white">Continue Reading</h2>
              <Link
                to="/library"
                className="text-primary font-semibold flex items-center gap-2 hover:text-primary/80 transition-colors group text-sm"
              >
                View All
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </Link>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {continueReading.map((item, idx) => (
                <motion.div
                  key={item.manga_id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: idx * 0.1 }}
                  className="group relative bg-background-light dark:bg-[#211c27] rounded-2xl p-4 border border-zinc-200 dark:border-zinc-800 hover:border-primary/50 transition-all cursor-pointer"
                  onClick={() => window.location.href = `/manga/${item.manga_id}`}
                >
                  <div className="flex gap-4">
                    <div className="flex-shrink-0 w-16 h-24 rounded-lg overflow-hidden bg-zinc-100 dark:bg-zinc-800">
                      {item.manga?.cover_url && (
                        <img
                          src={item.manga.cover_url}
                          alt={item.manga.title}
                          className="w-full h-full object-cover"
                          onError={(e) => { e.target.style.display = 'none'; }}
                        />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-bold text-zinc-900 dark:text-white mb-1 truncate">
                        {item.manga?.title || 'Unknown Manga'}
                      </h3>
                      <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-2">
                        Chapter {item.current_chapter || 0}
                      </p>
                      <div className="w-full bg-zinc-200 dark:bg-zinc-700 rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full transition-all"
                          style={{
                            width: `${item.manga?.total_chapters ? (item.current_chapter / item.manga.total_chapters) * 100 : 0}%`
                          }}
                        ></div>
                      </div>
                    </div>
                    <ArrowRight className="w-5 h-5 text-zinc-400 group-hover:text-primary group-hover:translate-x-1 transition-all flex-shrink-0 mt-1" />
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Trending Now Section */}
      <section className="py-12 bg-background-light dark:bg-background-dark">
        <div className="container mx-auto px-4">
          <div className="flex items-end justify-between mb-8">
            <div>
              <h2 className="text-3xl font-black text-zinc-900 dark:text-white mb-2">Trending Now</h2>
            </div>
            <Link
              to="/browse?sort=popular"
              className="text-primary font-semibold flex items-center gap-2 hover:text-primary/80 transition-colors group text-sm"
            >
              View All
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </Link>
          </div>

          {loading ? (
            <LoadingSpinner message="Loading trending manga..." />
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-6">
              {trendingManga.map((manga, idx) => (
                <MangaCard key={manga.id} manga={manga} index={idx} />
              ))}
            </div>
          )}
        </div>
      </section>

      {/* Recently Added Section */}
      <section className="py-12 bg-white dark:bg-[#191022]">
        <div className="container mx-auto px-4">
          <div className="flex items-end justify-between mb-8">
            <div>
              <h2 className="text-3xl font-black text-zinc-900 dark:text-white mb-2">Recently Added</h2>
            </div>
            <Link
              to="/browse?sort=year"
              className="text-primary font-semibold flex items-center gap-2 hover:text-primary/80 transition-colors group text-sm"
            >
              View All
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </Link>
          </div>

          {loading ? (
            <LoadingSpinner message="Loading recently added manga..." />
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-6">
              {recentlyAdded.map((manga, idx) => (
                <MangaCard key={manga.id} manga={manga} index={idx} />
              ))}
            </div>
          )}
        </div>
      </section>
    </div>
  );
};

export default Home;