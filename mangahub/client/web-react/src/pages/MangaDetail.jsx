import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Book, ArrowLeft, Plus, Check, TrendingUp, AlertCircle, MessageCircle, Search, ArrowUpDown, ChevronLeft, ChevronRight, X, List, ArrowUpCircle, ArrowDownCircle, BookOpen, Star
} from 'lucide-react';
import mangaService from '../services/mangaService';
import libraryService from '../services/libraryService';
import ratingService from '../services/ratingService';
import authService from '../services/authService';
import LoadingSpinner from '../components/LoadingSpinner';
import AddToLibraryModal from '../components/AddToLibraryModal';

const MangaDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [manga, setManga] = useState(null);
  const fetchingChaptersRef = useRef(false);
  const fetchingRatingsRef = useRef(false);
  const [recommendations, setRecommendations] = useState([]);
  const [inLibrary, setInLibrary] = useState(false);
  const [libraryEntry, setLibraryEntry] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [updating, setUpdating] = useState(false);
  const [showIndexModal, setShowIndexModal] = useState(false);
  const [sortAscending, setSortAscending] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [realChapters, setRealChapters] = useState([]);
  const [loadingChapters, setLoadingChapters] = useState(false);
  const [useRealChapters, setUseRealChapters] = useState(true);
  const [ratingStats, setRatingStats] = useState(null);
  const [userRating, setUserRating] = useState(null);
  const [hoverRating, setHoverRating] = useState(0);
  const [submittingRating, setSubmittingRating] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(authService.isAuthenticated());
  const [showAddModal, setShowAddModal] = useState(false);

  // Update authentication state whenever component updates or page is focused
  useEffect(() => {
    const checkAuth = () => {
      const authStatus = authService.isAuthenticated();
      setIsAuthenticated(authStatus);
    };
    
    checkAuth();
    
    // Check auth when window gains focus (user might have logged in another tab)
    window.addEventListener('focus', checkAuth);
    
    return () => {
      window.removeEventListener('focus', checkAuth);
    };
  }, []);

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
      if (isAuthenticated && authService.getToken()) {
        try {
          const libraryData = await libraryService.getLibrary();
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

  // Fetch real chapters from API
  useEffect(() => {
    const fetchChapters = async () => {
      if (!manga || fetchingChaptersRef.current) return;
      
      fetchingChaptersRef.current = true;
      // Fetch chapters for all manga (local database now has chapters)
      // IDs can be: "md-{uuid}" for MangaDex, "mal-{id}" for MAL, etc.
      const idStr = id.toString();

      try {
        setLoadingChapters(true);
        const chaptersData = await mangaService.getChapters(id, ['en'], 500, 0);
        if (chaptersData && chaptersData.chapters && chaptersData.chapters.length > 0) {
          setRealChapters(chaptersData.chapters);
          setUseRealChapters(true);
        } else {
          setUseRealChapters(false);
        }
      } catch (err) {
        console.error('Error fetching chapters:', err);
        setUseRealChapters(false);
      } finally {
        setLoadingChapters(false);
        fetchingChaptersRef.current = false;
      }
    };

    fetchChapters();
  }, [manga, id]);

  // Fetch rating stats
  useEffect(() => {
    const fetchRatings = async () => {
      if (!manga || fetchingRatingsRef.current) return;

      fetchingRatingsRef.current = true;
      try {
        const stats = await ratingService.getMangaRatings(id);
        console.log('📊 Rating stats received:', stats);
        setRatingStats(stats);
        setUserRating(stats.user_rating || null);
      } catch (err) {
        console.error('Error fetching ratings:', err);
      } finally {
        fetchingRatingsRef.current = false;
      }
    };

    fetchRatings();
  }, [manga, id]);

  const handleRating = async (rating) => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    try {
      setSubmittingRating(true);
      
      // If clicking the same star, unrate (delete rating)
      if (userRating === rating) {
        await ratingService.deleteRating(id);
        // Fetch updated stats after deletion
        const stats = await ratingService.getMangaRatings(id);
        setRatingStats(stats);
        // user_rating will be 0 if not rated, convert to null for frontend
        setUserRating(stats.user_rating > 0 ? stats.user_rating : null);
      } else {
        // Submit new rating
        const response = await ratingService.rateManga(id, rating);
        // Immediately set the user rating to what was just submitted
        setUserRating(rating);
        // Then fetch updated stats to get the new average and distribution
        const stats = await ratingService.getMangaRatings(id);
        setRatingStats(stats);
      }
    } catch (err) {
      console.error('Error rating manga:', err);
      alert('Failed to submit rating');
    } finally {
      setSubmittingRating(false);
    }
  };

  const handleAddToLibrary = async (status) => {
    if (!isAuthenticated || !authService.getToken()) {
      alert('Please log in to add manga to your library');
      navigate('/login');
      return;
    }

    setUpdating(true);
    try {
      await libraryService.addToLibrary(id, status); // Use id directly as string with selected status
      setInLibrary(true);
      setShowAddModal(false);
      await fetchMangaDetail(); // Refresh to get library entry
    } catch (err) {
      console.error('Error adding to library:', err);
      if (err.message && err.message.includes('No authentication token')) {
        alert('Your session has expired. Please log in again.');
        navigate('/login');
      } else {
        alert(err.error || err.message || 'Failed to add to library');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleRemoveFromLibrary = async () => {
    if (!isAuthenticated || !authService.getToken()) {
      alert('Please log in to manage your library');
      navigate('/login');
      return;
    }

    setUpdating(true);
    try {
      await libraryService.removeFromLibrary(id);
      setInLibrary(false);
      setLibraryEntry(null);
      setShowAddModal(false);
      await fetchMangaDetail(); // Refresh
    } catch (err) {
      console.error('Error removing from library:', err);
      alert(err.error || err.message || 'Failed to remove from library');
    } finally {
      setUpdating(false);
    }
  };

  const handleOpenAddModal = () => {
    if (!isAuthenticated || !authService.getToken()) {
      alert('Please log in to add manga to your library');
      navigate('/login');
      return;
    }
    setShowAddModal(true);
  };

  const handleUpdateProgress = async (currentChapter) => {
    setUpdating(true);
    try {
      await libraryService.updateProgress(parseInt(id), currentChapter, libraryEntry?.status || 'reading');
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
      await libraryService.updateProgress(parseInt(id), libraryEntry?.current_chapter || 0, newStatus);
      await fetchMangaDetail(); // Refresh library entry
    } catch (err) {
      console.error('Error updating status:', err);
      alert(err.message);
    } finally {
      setUpdating(false);
    }
  };

  // Generate chapters with volume information
  const chapters = useMemo(() => {
    if (!manga) return [];
    // Use real chapters if available, otherwise generate from total_chapters
    if (useRealChapters && realChapters.length > 0) {
      return realChapters.map(ch => ({
        id: ch.id,
        number: parseFloat(ch.chapter_number) || 0,
        volume: ch.volume_number ? parseInt(ch.volume_number) : null,
        title: ch.title || `Chapter ${ch.chapter_number}`,
        publishDate: ch.published_at || null,
        read: false,
        source: ch.source || 'mangadex',
        pages: ch.pages || 0,
        scanlation_group: ch.scanlation_group || 'Unknown',
        external_url: ch.external_url || null,
        is_external: ch.is_external || false
      }));
    }
    return mangaService.getChapterList(manga);
  }, [manga, useRealChapters, realChapters]);

  // Group chapters by volume
  const volumeGroups = useMemo(() => {
    const groups = {};
    chapters.forEach(chapter => {
      // Use 'no-volume' key for chapters without volume info instead of null
      const volumeKey = chapter.volume !== null && chapter.volume !== undefined ? chapter.volume : 'no-volume';
      if (!groups[volumeKey]) {
        groups[volumeKey] = [];
      }
      groups[volumeKey].push(chapter);
    });
    return groups;
  }, [chapters]);

  // Check if we have meaningful volumes (more than just no-volume group)
  const hasVolumes = useMemo(() => {
    const keys = Object.keys(volumeGroups);
    // Has volumes if there are volume keys other than 'no-volume', or if only 'no-volume' exists
    return keys.length > 0 && (keys.some(k => k !== 'no-volume') || keys.length === 1);
  }, [volumeGroups]);
  const itemsPerPage = hasVolumes ? 1 : 20; // 1 volume per page or 20 chapters per page

  // Get paginated content
  const paginatedContent = useMemo(() => {
    if (hasVolumes) {
      const volumes = Object.keys(volumeGroups)
        .map(v => v === 'no-volume' ? -1 : Number(v))
        .sort((a, b) => sortAscending ? a - b : b - a);
      const currentVolumeNum = volumes[currentPage - 1];
      if (currentVolumeNum === undefined) return [];
      const currentVolume = currentVolumeNum === -1 ? 'no-volume' : currentVolumeNum;
      const volumeChapters = volumeGroups[currentVolume];
      if (!volumeChapters || volumeChapters.length === 0) return [];
      return sortAscending ? volumeChapters : [...volumeChapters].reverse();
    } else {
      const startIdx = (currentPage - 1) * itemsPerPage;
      const endIdx = startIdx + itemsPerPage;
      const sortedChapters = sortAscending ? [...chapters] : [...chapters].reverse();
      return sortedChapters.slice(startIdx, endIdx);
    }
  }, [chapters, volumeGroups, hasVolumes, currentPage, itemsPerPage, sortAscending]);

  const totalPages = useMemo(() => {
    if (hasVolumes) {
      return Object.keys(volumeGroups).length;
    }
    return Math.ceil(chapters.length / itemsPerPage);
  }, [chapters.length, hasVolumes, volumeGroups, itemsPerPage]);

  const currentVolume = useMemo(() => {
    if (hasVolumes) {
      const volumes = Object.keys(volumeGroups)
        .map(v => v === 'no-volume' ? -1 : Number(v))
        .sort((a, b) => sortAscending ? a - b : b - a);
      const volumeNum = volumes[currentPage - 1];
      return volumeNum === -1 ? 'no-volume' : volumeNum;
    }
    return null;
  }, [hasVolumes, volumeGroups, currentPage, sortAscending]);

  const handlePageChange = (newPage) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setCurrentPage(newPage);
    }
  };

  const handleVolumeSelect = (volume) => {
    const volumes = Object.keys(volumeGroups)
      .map(v => v === 'no-volume' ? -1 : Number(v))
      .sort((a, b) => sortAscending ? a - b : b - a);
    const volumeNum = volume === 'no-volume' ? -1 : Number(volume);
    const pageIndex = volumes.indexOf(volumeNum);
    if (pageIndex !== -1) {
      setCurrentPage(pageIndex + 1);
      setShowIndexModal(false);
    }
  };

  const toggleSort = () => {
    setSortAscending(!sortAscending);
  };

  const handleChapterClick = (chapter) => {
    // Check if this is an external chapter (licensed manga)
    if (chapter.is_external && chapter.external_url) {
      // Open external URL in new tab
      window.open(chapter.external_url, '_blank', 'noopener,noreferrer');
      return;
    }

    if (useRealChapters && chapter.id) {
      // Navigate to reader with real chapter data
      navigate(`/read/${id}?chapter=${chapter.id}&source=${chapter.source}&number=${chapter.number}`);
    } else {
      // For generated chapters, show info message
      alert('This manga uses MAL for metadata. We will search MangaDex automatically when you click "Read". If chapters are not found, they may not be available on MangaDex yet.');
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
                    onClick={handleOpenAddModal}
                    disabled={updating || !isAuthenticated}
                    className="w-full flex items-center justify-center gap-2 px-6 py-3.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition font-bold shadow-lg shadow-primary/25 disabled:opacity-50"
                  >
                    <Plus className="w-5 h-5" />
                    <span>{isAuthenticated ? 'Add to Library' : 'Login to Add'}</span>
                  </button>
                ) : (
                  <button 
                    onClick={handleOpenAddModal}
                    disabled={updating}
                    className="w-full flex items-center justify-center gap-2 px-6 py-3.5 bg-green-600 text-white rounded-xl hover:bg-green-700 transition font-bold disabled:opacity-50"
                  >
                    <Check className="w-5 h-5" />
                    <span>{libraryEntry?.status ? libraryEntry.status.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase()) : 'In Library'}</span>
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

            {/* Rating Section */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl p-8 border border-zinc-200 dark:border-zinc-800">
              <h2 className="text-xl font-bold text-zinc-900 dark:text-white mb-6">Rating</h2>
              
              {ratingStats && (
                <div className="mb-8">
                  {/* Rating Summary and Distribution */}
                  <div className="flex flex-col md:flex-row items-start gap-8">
                    {/* Left side - Average Rating */}
                    <div className="flex flex-col items-center justify-center min-w-[140px]">
                      <div className="text-6xl font-black text-zinc-900 dark:text-white mb-3">
                        {ratingStats.average_rating ? ratingStats.average_rating.toFixed(1) : '0.0'}
                      </div>
                      <div className="flex gap-1 mb-2">
                        {[...Array(5)].map((_, i) => (
                          <Star
                            key={i}
                            className={`w-5 h-5 ${
                              i < Math.round(ratingStats.average_rating || 0)
                                ? 'fill-yellow-400 text-yellow-400'
                                : 'text-zinc-300 dark:text-zinc-700'
                            }`}
                          />
                        ))}
                      </div>
                      <div className="text-sm text-zinc-600 dark:text-zinc-400">
                        {ratingStats.total_ratings.toLocaleString()} rating{ratingStats.total_ratings !== 1 ? 's' : ''}
                      </div>
                    </div>

                    {/* Right side - Rating Distribution */}
                    <div className="flex-1 w-full md:w-auto space-y-2 min-w-[250px]">
                      {[5, 4, 3, 2, 1].map((star) => {
                        const dist = ratingStats.rating_distribution || {};
                        const count = dist[star] || dist[star.toString()] || 0;
                        const percentage = ratingStats.total_ratings > 0 
                          ? (count / ratingStats.total_ratings) * 100 
                          : 0;
                        
                        return (
                          <div key={star} className="flex items-center gap-3">
                            <span className="text-sm font-bold text-zinc-700 dark:text-zinc-300 w-4">
                              {star}
                            </span>
                            <div className="flex-1 bg-zinc-200 dark:bg-zinc-800 rounded-full h-3.5 overflow-hidden">
                              <div
                                className="bg-green-500 h-full rounded-full transition-all duration-300"
                                style={{ width: `${percentage}%`, minWidth: percentage > 0 ? '2%' : '0%' }}
                              />
                            </div>
                            <span className="text-sm text-zinc-600 dark:text-zinc-400 w-12 text-right font-medium">
                              {count.toLocaleString()}
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>
              )}

              {isAuthenticated ? (
                <div>
                  <h3 className="text-sm font-semibold text-zinc-700 dark:text-zinc-300 mb-3 text-center">
                    {userRating ? 'Your Rating' : 'Rate this manga'}
                  </h3>
                  {userRating && (
                    <p className="text-sm text-zinc-600 dark:text-zinc-400 text-center mb-3">
                      You rated this manga {userRating}/5
                    </p>
                  )}
                  <div className="flex justify-center gap-2">
                    {[...Array(5)].map((_, i) => {
                      const rating = i + 1;
                      const isHovered = hoverRating > 0 && rating <= hoverRating;
                      const isSelected = userRating && rating <= userRating;
                      const shouldHighlight = isHovered || (isSelected && hoverRating === 0);

                      return (
                        <button
                          key={i}
                          onClick={() => handleRating(rating)}
                          onMouseEnter={() => setHoverRating(rating)}
                          onMouseLeave={() => setHoverRating(0)}
                          disabled={submittingRating}
                          className="transition-all duration-200 hover:scale-110 disabled:opacity-50"
                          title={userRating === rating ? 'Click to remove rating' : `Rate ${rating}/5`}
                        >
                          <Star
                            className={`w-10 h-10 ${
                              shouldHighlight
                                ? 'fill-yellow-400 text-yellow-400'
                                : 'text-zinc-300 dark:text-zinc-700'
                            } transition-colors`}
                          />
                        </button>
                      );
                    })}
                  </div>
                </div>
              ) : (
                <div className="text-center py-4">
                  <p className="text-zinc-600 dark:text-zinc-400 mb-3">
                    Sign in to rate this manga
                  </p>
                  <button
                    onClick={() => navigate('/login')}
                    className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition"
                  >
                    Sign In
                  </button>
                </div>
              )}
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
                <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
                  Chapters {hasVolumes && currentVolume && `- Volume ${currentVolume}`}
                </h2>
                <div className="flex gap-2">
                  <button 
                    onClick={() => setShowIndexModal(true)}
                    className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition"
                    title="Index"
                  >
                    <List className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                  </button>
                  <button 
                    onClick={toggleSort}
                    className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition"
                    title={sortAscending ? "Sort Descending" : "Sort Ascending"}
                  >
                    {sortAscending ? (
                      <ArrowUpCircle className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                    ) : (
                      <ArrowDownCircle className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                    )}
                  </button>
                </div>
              </div>

              {/* Info Notice for MAL manga without chapters */}
              {!useRealChapters && !loadingChapters && (
                <div className="mb-4 p-4 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg">
                  <div className="flex items-start gap-3">
                    <BookOpen className="w-5 h-5 text-amber-600 dark:text-amber-400 mt-0.5 flex-shrink-0" />
                    <div className="text-sm text-amber-800 dark:text-amber-200">
                      <p className="font-semibold mb-1">No chapters found</p>
                      <p>This manga is not currently available for reading on MangaDex. Chapter availability depends on scanlation groups uploading to MangaDex.</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Info Notice for successfully loaded chapters */}
              {useRealChapters && realChapters.length > 0 && (
                <div className="mb-4 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
                  <div className="flex items-start gap-3">
                    <BookOpen className="w-5 h-5 text-green-600 dark:text-green-400 mt-0.5 flex-shrink-0" />
                    <div className="text-sm text-green-800 dark:text-green-200">
                      <p className="font-semibold mb-1">Chapters available from MangaDex</p>
                      <p>Found {realChapters.length} chapter{realChapters.length !== 1 ? 's' : ''} available for reading.</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Display chapters if we have real chapters OR if manga has total_chapters from MAL */}
              {(useRealChapters && realChapters.length > 0) || (manga.total_chapters && manga.total_chapters > 0) ? (
                <>
                  {loadingChapters && (
                    <div className="text-center py-4">
                      <div className="inline-flex items-center gap-2 text-zinc-500 dark:text-zinc-400">
                        <div className="animate-spin rounded-full h-4 w-4 border-2 border-primary border-t-transparent"></div>
                        <span className="text-sm">Loading chapters...</span>
                      </div>
                    </div>
                  )}
                  {paginatedContent.length > 0 && (
                    <div className="space-y-2 max-h-96 overflow-y-auto">
                    {paginatedContent.map((chapter) => {
                      const isRead = libraryEntry && libraryEntry.current_chapter >= chapter.number;
                      return (
                        <div
                          key={chapter.id || chapter.number}
                          onClick={() => handleChapterClick(chapter)}
                          className={`flex items-center justify-between p-4 rounded-lg border transition-colors cursor-pointer ${
                            isRead
                              ? 'bg-primary/5 border-primary/20 hover:bg-primary/10'
                              : 'bg-zinc-50 dark:bg-zinc-800/50 border-zinc-200 dark:border-zinc-700 hover:border-primary/50 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                          }`}
                        >
                          <div className="flex items-center gap-3">
                            {isRead && <Check className="w-5 h-5 text-primary" />}
                            <div>
                              <p className="font-semibold text-zinc-900 dark:text-white">
                                Chapter {chapter.number}
                                {chapter.title && chapter.title !== `Chapter ${chapter.number}` && (
                                  <span className="text-sm font-normal text-zinc-500 dark:text-zinc-400 ml-2">
                                    - {chapter.title}
                                  </span>
                                )}
                              </p>
                              <div className="flex items-center gap-2 text-sm text-zinc-500 dark:text-zinc-400">
                                <span>{chapter.publishDate || 'Publication date unknown'}</span>
                                {useRealChapters && chapter.pages > 0 && (
                                  <span>• {chapter.pages} pages</span>
                                )}
                                {useRealChapters && chapter.scanlation_group && chapter.scanlation_group !== 'Unknown' && (
                                  <span>• {chapter.scanlation_group}</span>
                                )}
                              </div>
                            </div>
                          </div>
                          {useRealChapters && (
                            <div className="flex items-center gap-2">
                              {chapter.is_external ? (
                                <span className="text-xs px-2 py-1 bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 rounded font-medium">
                                  {chapter.source === 'mangaplus' ? 'MangaPlus' : chapter.source === 'mangadex' ? 'MangaDex' : 'External'}
                                </span>
                              ) : (
                                <span className="text-xs px-2 py-1 bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 rounded font-medium">
                                  {chapter.source === 'mangadex' ? 'MangaDex' : chapter.source === 'mangaplus' ? 'MangaPlus' : chapter.source}
                                </span>
                              )}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                  )}

                  {/* Pagination Controls */}
                  {totalPages > 1 && (
                    <div className="flex items-center justify-between mt-6 pt-6 border-t border-zinc-200 dark:border-zinc-800">
                      <button
                        onClick={() => handlePageChange(currentPage - 1)}
                        disabled={currentPage === 1}
                        className="flex items-center gap-2 px-4 py-2 bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-white rounded-lg hover:bg-zinc-200 dark:hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
                      >
                        <ChevronLeft className="w-4 h-4" />
                        Previous
                      </button>
                      
                      <div className="text-sm text-zinc-600 dark:text-zinc-400">
                        {hasVolumes ? (
                          <span>
                            {currentVolume === 'no-volume' ? 'Chapters (No Volume)' : `Volume ${currentVolume}`} 
                          </span>
                        ) : (
                          <span>Page {currentPage} of {totalPages}</span>
                        )}
                      </div>

                      <button
                        onClick={() => handlePageChange(currentPage + 1)}
                        disabled={currentPage === totalPages}
                        className="flex items-center gap-2 px-4 py-2 bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-white rounded-lg hover:bg-zinc-200 dark:hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
                      >
                        Next
                        <ChevronRight className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </>
              ) : (
                <div className="text-center py-8">
                  <Book className="w-12 h-12 mx-auto mb-3 text-zinc-300 dark:text-zinc-700" />
                  <p className="text-zinc-500 dark:text-zinc-400">No chapter information available</p>
                </div>
              )}
            </div>
          </div>
        </motion.div>

        {/* Index Modal */}
        <AnimatePresence>
          {showIndexModal && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
              onClick={() => setShowIndexModal(false)}
            >
              <motion.div
                initial={{ scale: 0.9, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.9, opacity: 0 }}
                className="bg-white dark:bg-[#191022] rounded-2xl p-6 max-w-2xl w-full max-h-[80vh] overflow-hidden border border-zinc-200 dark:border-zinc-800"
                onClick={(e) => e.stopPropagation()}
              >
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-2xl font-bold text-zinc-900 dark:text-white">Index</h2>
                  <button
                    onClick={() => setShowIndexModal(false)}
                    className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition"
                  >
                    <X className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                  </button>
                </div>

                <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-4">
                  The Index ignores user blocks, group blocks, and language filters.
                </p>

                {hasVolumes ? (
                  <div className="overflow-y-auto max-h-[60vh] space-y-2">
                    {Object.keys(volumeGroups)
                      .map(v => v === 'no-volume' ? -1 : Number(v))
                      .sort((a, b) => a - b)
                      .map((volumeNum) => {
                        const volume = volumeNum === -1 ? 'no-volume' : volumeNum;
                        const volumeChapters = volumeGroups[volume];
                        if (!volumeChapters || volumeChapters.length === 0) return null;
                        
                        const chapterRange = `Chapter ${volumeChapters[0]?.number || '?'} - ${volumeChapters[volumeChapters.length - 1]?.number || '?'}`;
                        const chapterCount = volumeChapters.length;
                        const isCurrentVolume = volume === currentVolume;

                        return (
                          <button
                            key={volume}
                            onClick={() => handleVolumeSelect(volume)}
                            className={`w-full text-left p-4 rounded-lg border transition-all ${
                              isCurrentVolume
                                ? 'bg-primary/10 border-primary/50 ring-2 ring-primary/20'
                                : 'bg-zinc-50 dark:bg-zinc-800/50 border-zinc-200 dark:border-zinc-700 hover:border-primary/50'
                            }`}
                          >
                            <div className="flex items-center justify-between">
                              <div>
                                <p className="font-bold text-zinc-900 dark:text-white">
                                  {volume === 'no-volume' ? 'Chapters (No Volume)' : `Volume ${volume}`}
                                </p>
                                <p className="text-sm text-zinc-500 dark:text-zinc-400">
                                  {chapterRange} ({chapterCount})
                                </p>
                              </div>
                              {isCurrentVolume && (
                                <Check className="w-5 h-5 text-primary" />
                              )}
                            </div>
                          </button>
                        );
                      }).filter(Boolean)}
                  </div>
                ) : (
                  <div className="overflow-y-auto max-h-[60vh] space-y-1">
                    {[...Array(totalPages)].map((_, pageIdx) => {
                      const startChapter = pageIdx * itemsPerPage + 1;
                      const endChapter = Math.min((pageIdx + 1) * itemsPerPage, chapters.length);
                      const isCurrentPage = pageIdx + 1 === currentPage;

                      return (
                        <button
                          key={pageIdx}
                          onClick={() => {
                            setCurrentPage(pageIdx + 1);
                            setShowIndexModal(false);
                          }}
                          className={`w-full text-left p-3 rounded-lg border transition-all ${
                            isCurrentPage
                              ? 'bg-primary/10 border-primary/50 ring-2 ring-primary/20'
                              : 'bg-zinc-50 dark:bg-zinc-800/50 border-zinc-200 dark:border-zinc-700 hover:border-primary/50'
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="font-semibold text-zinc-900 dark:text-white">
                                Chapters {startChapter} - {endChapter}
                              </p>
                            </div>
                            {isCurrentPage && (
                              <Check className="w-5 h-5 text-primary" />
                            )}
                          </div>
                        </button>
                      );
                    })}
                  </div>
                )}
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Add to Library Modal */}
        <AddToLibraryModal
          isOpen={showAddModal}
          onClose={() => setShowAddModal(false)}
          manga={manga}
          onAdd={handleAddToLibrary}
          onRemove={handleRemoveFromLibrary}
          currentStatus={libraryEntry?.status}
        />
      </div>
    </div>
  );
};

export default MangaDetail;
