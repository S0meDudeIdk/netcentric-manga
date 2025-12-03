import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { 
  Book, Star, Calendar, User, Tag, ArrowLeft, Plus, Check, 
  TrendingUp, BookOpen, AlertCircle, MessageCircle 
} from 'lucide-react';
import mangaService from '../services/mangaService';
import userService from '../services/userService';
import authService from '../services/authService';
import LoadingSpinner from '../components/LoadingSpinner';

const MangaDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [manga, setManga] = useState(null);
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
      } else {
        // Try local database first
        console.log('Fetching local manga with ID:', id);
        data = await mangaService.getMangaById(id);
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
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <AlertCircle className="w-16 h-16 text-red-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Manga Not Found</h2>
          <p className="text-gray-600 mb-6">{error || 'This manga does not exist'}</p>
          <button
            onClick={() => navigate('/browse')}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Browse Manga
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        {/* Back Button */}
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          <span>Back</span>
        </button>

        {/* Manga Details */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white rounded-2xl shadow-xl overflow-hidden"
        >
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 p-8">
            {/* Cover Image */}
            <div className="md:col-span-1">
              <div className="relative rounded-lg overflow-hidden shadow-lg">
                {manga.cover_url ? (
                  <img
                    src={manga.cover_url}
                    alt={manga.title}
                    className="w-full h-auto object-cover"
                    onError={(e) => {
                      e.target.src = 'https://via.placeholder.com/300x400?text=No+Cover';
                    }}
                  />
                ) : (
                  <div className="w-full h-96 flex items-center justify-center bg-gradient-to-br from-blue-100 to-purple-100">
                    <Book className="w-24 h-24 text-gray-400" />
                  </div>
                )}

                {/* Rating Badge */}
                {manga.rating && (
                  <div className="absolute top-4 right-4 bg-yellow-400 text-white px-3 py-2 rounded-lg flex items-center gap-2 font-bold shadow-lg">
                    <Star className="w-5 h-5 fill-current" />
                    <span>{manga.rating.toFixed(1)}</span>
                  </div>
                )}
              </div>

              {/* Add to Library / Status */}
              <div className="mt-6 space-y-3">
                {/* ChatHub Button */}
                <button
                  onClick={() => navigate(`/chathub/${id}`)}
                  className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-lg hover:from-purple-700 hover:to-blue-700 transition font-semibold shadow-lg"
                >
                  <MessageCircle className="w-5 h-5" />
                  <span>Join Chat Hub</span>
                </button>

                {!inLibrary ? (
                  <button
                    onClick={handleAddToLibrary}
                    disabled={updating || !isAuthenticated}
                    className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition font-semibold disabled:opacity-50"
                  >
                    <Plus className="w-5 h-5" />
                    <span>{isAuthenticated ? 'Add to Library' : 'Login to Add'}</span>
                  </button>
                ) : (
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 px-4 py-3 bg-green-50 border border-green-200 rounded-lg">
                      <Check className="w-5 h-5 text-green-600" />
                      <span className="text-green-700 font-semibold">In Your Library</span>
                    </div>

                    {/* Status Selector */}
                    <select
                      value={libraryEntry?.status || 'plan_to_read'}
                      onChange={(e) => handleStatusChange(e.target.value)}
                      disabled={updating}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="reading">Reading</option>
                      <option value="completed">Completed</option>
                      <option value="plan_to_read">Plan to Read</option>
                      <option value="on_hold">On Hold</option>
                      <option value="dropped">Dropped</option>
                    </select>

                    {/* Progress */}
                    <div className="bg-gray-50 rounded-lg p-4">
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Current Chapter
                      </label>
                      <div className="flex gap-2">
                        <input
                          type="number"
                          min="0"
                          max={manga.total_chapters || 999}
                          defaultValue={libraryEntry?.current_chapter || 0}
                          onBlur={(e) => handleUpdateProgress(parseInt(e.target.value))}
                          className="flex-1 px-4 py-2 border border-gray-300 rounded-lg"
                        />
                        <span className="flex items-center px-3 text-gray-600">
                          / {manga.total_chapters || '?'}
                        </span>
                      </div>
                      <div className="mt-2 bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-blue-600 rounded-full h-2 transition-all"
                          style={{
                            width: `${Math.min(100, ((libraryEntry?.current_chapter || 0) / (manga.total_chapters || 1)) * 100)}%`
                          }}
                        />
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Info */}
            <div className="md:col-span-2">
              <h1 className="text-4xl font-bold text-gray-900 mb-4">{manga.title}</h1>

              {/* Meta Info */}
              <div className="flex flex-wrap gap-4 mb-6">
                <div className="flex items-center gap-2 text-gray-600">
                  <User className="w-5 h-5" />
                  <span>{manga.author}</span>
                </div>
                {manga.publication_year && (
                  <div className="flex items-center gap-2 text-gray-600">
                    <Calendar className="w-5 h-5" />
                    <span>{manga.publication_year}</span>
                  </div>
                )}
                <div className="flex items-center gap-2 text-gray-600">
                  <BookOpen className="w-5 h-5" />
                  <span>{manga.total_chapters || 0} chapters</span>
                </div>
              </div>

              {/* Status Badge */}
              {manga.status && (
                <div className="mb-6">
                  <span
                    className={`inline-block px-4 py-2 text-sm font-semibold rounded-lg ${
                      manga.status === 'completed'
                        ? 'bg-green-100 text-green-700'
                        : manga.status === 'ongoing'
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-gray-100 text-gray-700'
                    }`}
                  >
                    {manga.status.charAt(0).toUpperCase() + manga.status.slice(1)}
                  </span>
                </div>
              )}

              {/* Genres */}
              {manga.genres && manga.genres.length > 0 && (
                <div className="mb-6">
                  <div className="flex items-center gap-2 mb-3">
                    <Tag className="w-5 h-5 text-gray-600" />
                    <h3 className="font-semibold text-gray-900">Genres</h3>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {manga.genres.map((genre, index) => (
                      <span
                        key={index}
                        className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm font-medium"
                      >
                        {genre}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Description */}
              <div className="mb-6">
                <h3 className="font-semibold text-gray-900 mb-3 text-lg">Description</h3>
                <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">
                  {manga.description || 'No description available.'}
                </p>
              </div>

              {/* Additional Stats */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-8">
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <TrendingUp className="w-6 h-6 text-blue-600 mx-auto mb-2" />
                  <p className="text-2xl font-bold text-gray-900">{manga.rating?.toFixed(1) || 'N/A'}</p>
                  <p className="text-sm text-gray-600">Rating</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <Book className="w-6 h-6 text-purple-600 mx-auto mb-2" />
                  <p className="text-2xl font-bold text-gray-900">{manga.total_chapters || 0}</p>
                  <p className="text-sm text-gray-600">Chapters</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <Calendar className="w-6 h-6 text-green-600 mx-auto mb-2" />
                  <p className="text-2xl font-bold text-gray-900">{manga.publication_year || 'N/A'}</p>
                  <p className="text-sm text-gray-600">Year</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <Tag className="w-6 h-6 text-orange-600 mx-auto mb-2" />
                  <p className="text-2xl font-bold text-gray-900">{manga.genres?.length || 0}</p>
                  <p className="text-sm text-gray-600">Genres</p>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
};

export default MangaDetail;
