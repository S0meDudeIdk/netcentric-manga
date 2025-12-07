import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Book, ArrowLeft, Plus, Check, TrendingUp, AlertCircle, MessageCircle, Search
} from 'lucide-react';
import mangaService from '../services/mangaService';
import userService from '../services/userService';
import authService from '../services/authService';
import LoadingSpinner from '../components/LoadingSpinner';

const MangaDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [manga, setManga] = useState(null);
  const [recommendations, setRecommendations] = useState([]);
  const [inLibrary, setInLibrary] = useState(false);
  const [libraryEntry, setLibraryEntry] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [updating, setUpdating] = useState(false);
  const isAuthenticated = authService.isAuthenticated();

  const fetchMangaDetail = useCallback(async () => {
    try {
      setLoading(true);

      // Check if this is a MAL ID (format: mal-123)
      let data;
      const idStr = id.toString();

      if (idStr.startsWith('mal-')) {
        // Extract MAL ID and fetch from MAL API
        const malId = idStr.replace('mal-', '');
        console.log('Fetching MAL manga with ID:', malId);
        data = await mangaService.getMALMangaById(malId);
        
        // Fetch recommendations for MAL manga
        try {
          const recData = await mangaService.getMALRecommendations(malId);
          setRecommendations(recData.data || []);
        } catch (err) {
          console.error('Error fetching recommendations:', err);
          setRecommendations([]);
        }
      } else {
        // Try local database first
        console.log('Fetching local manga with ID:', id);
        data = await mangaService.getMangaById(id);
        setRecommendations([]);
      }

      console.log('Fetched manga data:', data);
      setManga(data);

      // Check if in library (only for authenticated users)
      if (isAuthenticated) {
        try {
          const libraryData = await userService.getLibrary();
          // For MAL manga, we need to match by the full ID including "mal-" prefix
          const searchId = idStr.startsWith('mal-') ? idStr : parseInt(id);
          const entry = libraryData.library?.find(item =>
            item.manga_id === searchId || item.manga_id === parseInt(id)
          );
          if (entry) {
            setInLibrary(true);
            setLibraryEntry(entry);
          }
        } catch (err) {
          console.error('Error checking library:', err);
        }
      }
    } catch (err) {
      console.error('Error fetching manga:', err);
      setError(err.response?.data?.error || err.message || 'Failed to load manga');
    } finally {
      setLoading(false);
    }
  }, [id, isAuthenticated]);

  useEffect(() => {
    fetchMangaDetail();
  }, [fetchMangaDetail]);

  const handleAddToLibrary = async () => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    setUpdating(true);
    try {
      await userService.addToLibrary(parseInt(id), 'plan_to_read');
      setInLibrary(true);
      await fetchMangaDetail(); // Refresh to get library entry
    } catch (err) {
      console.error('Error adding to library:', err);
      alert(err.message);
    } finally {
      setUpdating(false);
    }
  };

  const handleUpdateProgress = async (currentChapter) => {
    setUpdating(true);
    try {
      await userService.updateProgress(parseInt(id), currentChapter, libraryEntry?.status || 'reading');
      await fetchMangaDetail(); // Refresh library entry
    } catch (err) {
      console.error('Error updating progress:', err);
      alert(err.message);
    } finally {
      setUpdating(false);
    }
  };

  const handleStatusChange = async (newStatus) => {
    setUpdating(true);
    try {
      await userService.updateProgress(parseInt(id), libraryEntry?.current_chapter || 0, newStatus);
      await fetchMangaDetail(); // Refresh library entry
    } catch (err) {
      console.error('Error updating status:', err);
      alert(err.message);
    } finally {
      setUpdating(false);
    }
  };

  if (loading) {
    return <LoadingSpinner message="Loading manga details..." />;
  }

  if (error || !manga) {
    return (
      <div className="min-h-screen bg-background-light dark:bg-background-dark flex items-center justify-center">
        <div className="text-center p-8 bg-white dark:bg-[#191022] rounded-2xl shadow-xl border border-zinc-200 dark:border-zinc-800">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">Manga Not Found</h2>
          <p className="text-zinc-600 dark:text-zinc-400 mb-6">{error || 'This manga does not exist'}</p>
          <button
            onClick={() => navigate('/browse')}
            className="px-6 py-2 bg-primary text-white rounded-xl hover:bg-primary/90 transition shadow-lg shadow-primary/25"
          >
            Browse Manga
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark py-8 transition-colors">
      <div className="container mx-auto px-4">
        {/* Back Button */}
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white mb-8 transition-colors group"
        >
          <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          <span>Back</span>
        </button>

        {/* Manga Details */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="grid grid-cols-1 lg:grid-cols-3 gap-8"
        >
          {/* Left Column - Cover */}
          <div className="lg:col-span-1">
            <div className="bg-white dark:bg-[#191022] rounded-2xl p-6 border border-zinc-200 dark:border-zinc-800 sticky top-24">
              {/* Cover Image */}
              <div className="relative rounded-xl overflow-hidden shadow-2xl shadow-black/20 mb-6">
                {manga.cover_url ? (
                  <img
                    src={manga.cover_url}
                    alt={manga.title}
                    className="w-full h-auto max-h-96 object-contain mx-auto"
                    onError={(e) => {
                      e.target.src = 'https://via.placeholder.com/300x400?text=No+Cover';
                    }}
                  />
                ) : (
                  <div className="w-full h-96 flex items-center justify-center bg-zinc-100 dark:bg-zinc-800">
                    <Book className="w-24 h-24 text-zinc-300 dark:text-zinc-600" />
                  </div>
                )}
              </div>

              {/* Action Buttons */}
              <div className="space-y-3 mb-6">
                {!inLibrary ? (
                  <button
                    onClick={handleAddToLibrary}
                    disabled={updating || !isAuthenticated}
                    className="w-full flex items-center justify-center gap-2 px-6 py-3.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition font-bold shadow-lg shadow-primary/25 disabled:opacity-50"
                  >
                    <Plus className="w-5 h-5" />
                    <span>{isAuthenticated ? 'Add to Library' : 'Login to Add'}</span>
                  </button>
                ) : (
                  <button className="w-full flex items-center justify-center gap-2 px-6 py-3.5 bg-green-600 text-white rounded-xl font-bold">
                    <Check className="w-5 h-5" />
                    <span>In Library</span>
                  </button>
                )}

                <button
                  className="w-full flex items-center justify-center gap-2 px-6 py-3.5 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-white border border-zinc-200 dark:border-zinc-700 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-700 transition font-bold"
                  onClick={() => navigate(`/chathub/${id}`)}
                >
                  <MessageCircle className="w-5 h-5" />
                  <span>Join Chat Hub</span>
                </button>
              </div>

              {/* My Progress Section */}
              {inLibrary && libraryEntry && (
                <div className="mt-6 pt-6 border-t border-zinc-200 dark:border-zinc-800">
                  <h3 className="font-bold text-zinc-900 dark:text-white mb-4">MY PROGRESS</h3>
                  
                  <div className="space-y-3">
                    <div>
                      <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-2">Status</label>
                      <select
                        value={libraryEntry?.status || 'plan_to_read'}
                        onChange={(e) => handleStatusChange(e.target.value)}
                        disabled={updating}
                        className="w-full px-4 py-2.5 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg focus:ring-2 focus:ring-primary/50 text-zinc-900 dark:text-white outline-none text-sm"
                      >
                        <option value="reading">Reading</option>
                        <option value="completed">Completed</option>
                        <option value="plan_to_read">Plan to Read</option>
                        <option value="on_hold">On Hold</option>
                        <option value="dropped">Dropped</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-xs font-medium text-zinc-500 dark:text-zinc-400 mb-2">Last Chapter Read</label>
                      <div className="flex gap-2">
                        <input
                          type="number"
                          min="0"
                          max={manga.total_chapters || 999}
                          defaultValue={libraryEntry?.current_chapter || 0}
                          onBlur={(e) => handleUpdateProgress(parseInt(e.target.value))}
                          className="flex-1 px-4 py-2.5 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-lg text-zinc-900 dark:text-white focus:ring-2 focus:ring-primary/50 outline-none text-sm"
                        />
                        <div className="flex items-center px-4 bg-zinc-100 dark:bg-zinc-800 rounded-lg text-zinc-600 dark:text-zinc-400 font-medium text-sm">
                          / {manga.total_chapters || '179'}
                        </div>
                      </div>
                      <div className="mt-3 bg-zinc-200 dark:bg-zinc-800 rounded-full h-1.5 overflow-hidden">
                        <div
                          className="bg-primary rounded-full h-full transition-all duration-500"
                          style={{
                            width: `${Math.min(100, ((libraryEntry?.current_chapter || 0) / (manga.total_chapters || 1)) * 100)}%`
                          }}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Right Column - Info and Chapters */}
          <div className="lg:col-span-2 space-y-6">
            {/* Title and Author */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl p-8 border border-zinc-200 dark:border-zinc-800">
              <h1 className="text-4xl md:text-5xl font-black text-zinc-900 dark:text-white mb-3 leading-tight">{manga.title}</h1>
              <p className="text-zinc-600 dark:text-zinc-400 text-lg">By {manga.author || 'Unknown'}</p>
            </div>

            {/* Description */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl p-8 border border-zinc-200 dark:border-zinc-800">
              <h2 className="text-xl font-bold text-zinc-900 dark:text-white mb-4">Description</h2>
              <p className="text-zinc-600 dark:text-zinc-300 leading-relaxed text-justify">
                {manga.description || 'No description available.'}
              </p>
            </div>

            {/* Genres and Details - keeping original structure below this */}
            <div style={{display: 'none'}}>
            </div>

            {/* Genres and Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Genres */}
              <div className="bg-white dark:bg-[#191022] rounded-2xl p-6 border border-zinc-200 dark:border-zinc-800">
                <h3 className="text-lg font-bold text-zinc-900 dark:text-white mb-4">Genres</h3>
                <div className="flex flex-wrap gap-2">
                  {manga.genres && manga.genres.length > 0 ? (
                    manga.genres.map((genre, index) => (
                      <span
                        key={index}
                        className="px-3 py-1.5 bg-primary/10 text-primary rounded-lg text-sm font-semibold"
                      >
                        {genre}
                      </span>
                    ))
                  ) : (
                    <span className="text-zinc-500 dark:text-zinc-400 text-sm">No genres available</span>
                  )}
                </div>
              </div>

              {/* Details */}
              <div className="bg-white dark:bg-[#191022] rounded-2xl p-6 border border-zinc-200 dark:border-zinc-800">
                <h3 className="text-lg font-bold text-zinc-900 dark:text-white mb-4">Details</h3>
                <div className="space-y-3 text-sm">
                  <div className="flex justify-between">
                    <span className="text-zinc-500 dark:text-zinc-400">Status:</span>
                    <span className={`font-semibold ${
                      manga.status === 'completed' 
                        ? 'text-green-600 dark:text-green-400' 
                        : 'text-blue-600 dark:text-blue-400'
                    }`}>
                      {manga.status ? manga.status.charAt(0).toUpperCase() + manga.status.slice(1) : 'Unknown'}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500 dark:text-zinc-400">Chapters:</span>
                    <span className="text-zinc-900 dark:text-white font-semibold">{manga.total_chapters || 'N/A'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500 dark:text-zinc-400">Published:</span>
                    <span className="text-zinc-900 dark:text-white font-semibold">{manga.publication_year ? `Jul 25, ${manga.publication_year} to Dec 25, 2021` : 'Unknown'}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Chapters List */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl p-8 border border-zinc-200 dark:border-zinc-800">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-zinc-900 dark:text-white">Chapters</h2>
                <div className="flex gap-2">
                  <button className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition">
                    <Search className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                  </button>
                  <button className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition">
                    <TrendingUp className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                  </button>
                </div>
              </div>

              {manga.total_chapters && manga.total_chapters > 0 ? (
                <div className="space-y-2 max-h-96 overflow-y-auto">
                  {[...Array(Math.min(10, manga.total_chapters))].map((_, idx) => {
                    const chapterNum = manga.total_chapters - idx;
                    const isRead = libraryEntry && libraryEntry.current_chapter >= chapterNum;
                    return (
                      <div
                        key={idx}
                        className={`flex items-center justify-between p-4 rounded-lg border transition-colors cursor-pointer ${
                          isRead
                            ? 'bg-primary/5 border-primary/20'
                            : 'bg-zinc-50 dark:bg-zinc-800/50 border-zinc-200 dark:border-zinc-700 hover:border-primary/50'
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          {isRead && <Check className="w-5 h-5 text-primary" />}
                          <div>
                            <p className="font-semibold text-zinc-900 dark:text-white">
                              Chapter {chapterNum}
                            </p>
                            <p className="text-sm text-zinc-500 dark:text-zinc-400">
                              {manga.publication_year ? `Published ${manga.publication_year}` : 'Publication date unknown'}
                            </p>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                  {manga.total_chapters > 10 && (
                    <p className="text-center text-sm text-zinc-500 dark:text-zinc-400 pt-4">
                      Showing latest 10 of {manga.total_chapters} chapters
                    </p>
                  )}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Book className="w-12 h-12 mx-auto mb-3 text-zinc-300 dark:text-zinc-700" />
                  <p className="text-zinc-500 dark:text-zinc-400">No chapter information available</p>
                </div>
              )}
            </div>

            {/* Recommended */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl p-8 border border-zinc-200 dark:border-zinc-800">
              <h2 className="text-xl font-bold text-zinc-900 dark:text-white mb-6">Recommended</h2>
              {recommendations && recommendations.length > 0 ? (
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4">
                  {recommendations.slice(0, 6).map((rec) => (
                    <div 
                      key={rec.id} 
                      className="group cursor-pointer"
                      onClick={() => navigate(`/manga/${rec.id}`)}
                    >
                      <div className="aspect-[2/3] bg-zinc-200 dark:bg-zinc-800 rounded-lg mb-2 overflow-hidden">
                        {rec.cover_url ? (
                          <img
                            src={rec.cover_url}
                            alt={rec.title}
                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                            onError={(e) => {
                              e.target.style.display = 'none';
                              e.target.parentElement.innerHTML = '<div class="w-full h-full flex items-center justify-center"><svg class="w-8 h-8 text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path></svg></div>';
                            }}
                          />
                        ) : (
                          <div className="w-full h-full flex items-center justify-center">
                            <Book className="w-8 h-8 text-zinc-400" />
                          </div>
                        )}
                      </div>
                      <p className="text-sm font-semibold text-zinc-900 dark:text-white truncate group-hover:text-primary transition">
                        {rec.title}
                      </p>
                      <div className="flex items-center gap-1 mt-1">
                        <span className="text-xs text-zinc-500 dark:text-zinc-400">
                          ⭐ {rec.rating ? rec.rating.toFixed(1) : 'N/A'}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Book className="w-12 h-12 mx-auto mb-3 text-zinc-300 dark:text-zinc-700" />
                  <p className="text-zinc-500 dark:text-zinc-400">No recommendations available</p>
                </div>
              )}
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
};

export default MangaDetail;
