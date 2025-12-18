import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  ArrowLeft, ChevronLeft, ChevronRight, Settings, X, AlertCircle
} from 'lucide-react';
import mangaService from '../services/mangaService';
import userService from '../services/userService';
import authService from '../services/authService';
import LoadingSpinner from '../components/LoadingSpinner';

const ChapterReader = () => {
  const { mangaId } = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const chapterId = searchParams.get('chapter');
  const source = searchParams.get('source') || 'mangadex';
  const chapterNumber = searchParams.get('number') || '';
  
  const [pages, setPages] = useState([]);
  const [currentPage, setCurrentPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showSettings, setShowSettings] = useState(false);
  const [chapters, setChapters] = useState([]);
  const [loadingChapters, setLoadingChapters] = useState(false);
  
  // Reader settings
  const [viewMode, setViewMode] = useState('single'); // 'single', 'double', 'vertical'
  const [fitMode, setFitMode] = useState('width'); // 'width', 'height', 'original'
  const [backgroundColor, setBackgroundColor] = useState('#1a1a1a');
  
  const imageRefs = useRef([]);
  const isAuthenticated = authService.isAuthenticated();

  // Fetch chapter pages
  const fetchChapterPages = useCallback(async () => {
    if (!chapterId) {
      setError('No chapter ID provided');
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const data = await mangaService.getChapterPages(chapterId, source);
      setPages(data.pages || []);
      setCurrentPage(0);
      
      // Update reading progress if authenticated
      if (isAuthenticated && chapterNumber) {
        try {
          await userService.updateProgress(mangaId, parseInt(chapterNumber), 'reading');
        } catch (err) {
          console.error('Failed to update progress:', err);
        }
      }
    } catch (err) {
      console.error('Error loading chapter:', err);
      setError(err.error || err.message || 'Failed to load chapter pages');
    } finally {
      setLoading(false);
    }
  }, [chapterId, source, isAuthenticated, mangaId, chapterNumber]);

  // Fetch chapter list
  const fetchChapterList = useCallback(async () => {
    try {
      setLoadingChapters(true);
      const data = await mangaService.getChapters(mangaId, ['en'], 500, 0);
      setChapters(data.chapters || []);
    } catch (err) {
      console.error('Error loading chapter list:', err);
    } finally {
      setLoadingChapters(false);
    }
  }, [mangaId]);

  useEffect(() => {
    fetchChapterPages();
    fetchChapterList();
  }, [fetchChapterPages, fetchChapterList]);

  // Navigation functions
  const handlePrevPage = useCallback(() => {
    if (currentPage > 0) {
      setCurrentPage(currentPage - 1);
      scrollToTop();
    }
  }, [currentPage]);

  const handleNextPage = useCallback(() => {
    if (currentPage < pages.length - 1) {
      setCurrentPage(currentPage + 1);
      scrollToTop();
    }
  }, [currentPage, pages.length]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyPress = (e) => {
      if (e.key === 'ArrowLeft' || e.key === 'a') {
        handlePrevPage();
      } else if (e.key === 'ArrowRight' || e.key === 'd') {
        handleNextPage();
      } else if (e.key === 'Escape') {
        setShowSettings(false);
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [handleNextPage, handlePrevPage]);

  const scrollToTop = () => {
    if (viewMode !== 'vertical') {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const handleNextChapter = () => {
    if (chapters.length === 0) return;
    
    const currentIndex = chapters.findIndex(ch => ch.id === chapterId);
    if (currentIndex !== -1 && currentIndex < chapters.length - 1) {
      const nextChapter = chapters[currentIndex + 1];
      navigate(`/read/${mangaId}?chapter=${nextChapter.id}&source=${nextChapter.source}&number=${nextChapter.chapter_number}`);
    }
  };

  const handlePrevChapter = () => {
    if (chapters.length === 0) return;
    
    const currentIndex = chapters.findIndex(ch => ch.id === chapterId);
    if (currentIndex > 0) {
      const prevChapter = chapters[currentIndex - 1];
      navigate(`/read/${mangaId}?chapter=${prevChapter.id}&source=${prevChapter.source}&number=${prevChapter.chapter_number}`);
    }
  };

  const getImageClass = () => {
    const baseClass = 'mx-auto block';
    if (fitMode === 'width') return `${baseClass} w-full h-auto`;
    if (fitMode === 'height') return `${baseClass} h-screen w-auto`;
    return `${baseClass}`;
  };

  if (loading) {
    return <LoadingSpinner message="Loading chapter..." />;
  }

  if (error || pages.length === 0) {
    return (
      <div className="min-h-screen bg-background-dark flex items-center justify-center p-4">
        <div className="text-center p-8 bg-[#191022] rounded-2xl shadow-xl border border-zinc-800 max-w-md">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-white mb-2">Failed to Load Chapter</h2>
          <p className="text-zinc-400 mb-6">{error || 'No pages available for this chapter'}</p>
          <button
            onClick={() => navigate(`/manga/${mangaId}`)}
            className="px-6 py-2 bg-primary text-white rounded-xl hover:bg-primary/90 transition"
          >
            Back to Manga Details
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen transition-colors" style={{ backgroundColor }}>
      {/* Header */}
      <div className="fixed top-0 left-0 right-0 z-50 bg-black/80 backdrop-blur-sm border-b border-zinc-800">
        <div className="container mx-auto px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button
                onClick={() => navigate(`/manga/${mangaId}`)}
                className="flex items-center gap-2 text-zinc-300 hover:text-white transition"
              >
                <ArrowLeft className="w-5 h-5" />
                <span className="hidden sm:inline">Back</span>
              </button>
              
              <div className="text-white">
                <p className="text-sm text-zinc-400">Chapter {chapterNumber}</p>
                <p className="text-xs text-zinc-500">
                  Page {currentPage + 1} / {pages.length}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={handlePrevChapter}
                disabled={loadingChapters}
                className="px-3 py-1.5 text-sm bg-zinc-800 text-white rounded-lg hover:bg-zinc-700 transition disabled:opacity-50"
              >
                Prev Ch
              </button>
              <button
                onClick={handleNextChapter}
                disabled={loadingChapters}
                className="px-3 py-1.5 text-sm bg-zinc-800 text-white rounded-lg hover:bg-zinc-700 transition disabled:opacity-50"
              >
                Next Ch
              </button>
              <button
                onClick={() => setShowSettings(!showSettings)}
                className="p-2 bg-zinc-800 text-white rounded-lg hover:bg-zinc-700 transition"
              >
                <Settings className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Settings Panel */}
      {showSettings && (
        <motion.div
          initial={{ opacity: 0, x: 300 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: 300 }}
          className="fixed top-16 right-4 z-50 bg-[#191022] rounded-2xl shadow-2xl border border-zinc-800 p-6 w-80"
        >
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-bold text-white">Reader Settings</h3>
            <button
              onClick={() => setShowSettings(false)}
              className="p-1 hover:bg-zinc-800 rounded-lg transition"
            >
              <X className="w-5 h-5 text-zinc-400" />
            </button>
          </div>

          <div className="space-y-4">
            {/* View Mode */}
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-2">View Mode</label>
              <select
                value={viewMode}
                onChange={(e) => setViewMode(e.target.value)}
                className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-white outline-none focus:ring-2 focus:ring-primary/50"
              >
                <option value="single">Single Page</option>
                <option value="vertical">Vertical Scroll</option>
              </select>
            </div>

            {/* Fit Mode */}
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-2">Fit Mode</label>
              <select
                value={fitMode}
                onChange={(e) => setFitMode(e.target.value)}
                className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-white outline-none focus:ring-2 focus:ring-primary/50"
              >
                <option value="width">Fit Width</option>
                <option value="height">Fit Height</option>
                <option value="original">Original Size</option>
              </select>
            </div>

            {/* Background Color */}
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-2">Background</label>
              <div className="grid grid-cols-4 gap-2">
                {['#000000', '#1a1a1a', '#2d2d2d', '#ffffff'].map((color) => (
                  <button
                    key={color}
                    onClick={() => setBackgroundColor(color)}
                    className={`w-full h-10 rounded-lg border-2 transition ${
                      backgroundColor === color ? 'border-primary' : 'border-zinc-700'
                    }`}
                    style={{ backgroundColor: color }}
                  />
                ))}
              </div>
            </div>
          </div>
        </motion.div>
      )}

      {/* Reader Content */}
      <div className="pt-20 pb-8">
        {viewMode === 'vertical' ? (
          // Vertical scroll mode - all pages
          <div className="space-y-1">
            {pages.map((pageUrl, index) => (
              <div key={index} className="relative">
                <img
                  src={pageUrl}
                  alt={`Page ${index + 1}`}
                  className={getImageClass()}
                  loading="lazy"
                  onError={(e) => {
                    e.target.src = 'https://via.placeholder.com/800x1200?text=Failed+to+Load';
                  }}
                />
              </div>
            ))}
          </div>
        ) : (
          // Single page mode
          <div className="relative">
            <img
              ref={(el) => (imageRefs.current[currentPage] = el)}
              src={pages[currentPage]}
              alt={`Page ${currentPage + 1}`}
              className={getImageClass()}
              onError={(e) => {
                e.target.src = 'https://via.placeholder.com/800x1200?text=Failed+to+Load';
              }}
            />
          </div>
        )}
      </div>

      {/* Navigation Controls (Single Page Mode) */}
      {viewMode === 'single' && (
        <div className="fixed bottom-8 left-1/2 transform -translate-x-1/2 z-40">
          <div className="flex items-center gap-4 bg-black/80 backdrop-blur-sm px-6 py-3 rounded-full border border-zinc-700">
            <button
              onClick={handlePrevPage}
              disabled={currentPage === 0}
              className="p-2 text-white hover:bg-zinc-800 rounded-full transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="w-6 h-6" />
            </button>

            <span className="text-white font-medium min-w-[100px] text-center">
              {currentPage + 1} / {pages.length}
            </span>

            <button
              onClick={handleNextPage}
              disabled={currentPage === pages.length - 1}
              className="p-2 text-white hover:bg-zinc-800 rounded-full transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="w-6 h-6" />
            </button>
          </div>
        </div>
      )}

      {/* Click Areas for Navigation (Single Page Mode) */}
      {viewMode === 'single' && (
        <>
          <div
            onClick={handlePrevPage}
            className="fixed left-0 top-20 bottom-0 w-1/3 cursor-pointer z-30"
            style={{ cursor: currentPage === 0 ? 'not-allowed' : 'pointer' }}
          />
          <div
            onClick={handleNextPage}
            className="fixed right-0 top-20 bottom-0 w-1/3 cursor-pointer z-30"
            style={{ cursor: currentPage === pages.length - 1 ? 'not-allowed' : 'pointer' }}
          />
        </>
      )}
    </div>
  );
};

export default ChapterReader;
