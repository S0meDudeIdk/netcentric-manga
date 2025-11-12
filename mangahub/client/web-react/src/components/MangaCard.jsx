import React from 'react';
import { motion } from 'framer-motion';
import { Book, Star, Calendar } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const MangaCard = ({ manga }) => {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/manga/${manga.id}`);
  };

  return (
    <motion.div
      whileHover={{ scale: 1.03, y: -5 }}
      whileTap={{ scale: 0.98 }}
      className="bg-white rounded-lg shadow-md overflow-hidden cursor-pointer hover:shadow-xl transition-shadow"
      onClick={handleClick}
    >
      {/* Cover Image */}
      <div className="relative h-64 bg-gray-200">
        {manga.cover_url ? (
          <img
            src={manga.cover_url}
            alt={manga.title}
            className="w-full h-full object-cover"
            onError={(e) => {
              e.target.src = 'https://via.placeholder.com/300x400?text=No+Cover';
            }}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-blue-100 to-purple-100">
            <Book className="w-16 h-16 text-gray-400" />
          </div>
        )}
        
        {/* Rating Badge */}
        {manga.rating && (
          <div className="absolute top-2 right-2 bg-yellow-400 text-white px-2 py-1 rounded-md flex items-center gap-1 text-sm font-semibold">
            <Star className="w-4 h-4 fill-current" />
            <span>{manga.rating.toFixed(1)}</span>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-4">
        <h3 className="font-bold text-lg mb-1 truncate" title={manga.title}>
          {manga.title}
        </h3>
        
        <p className="text-gray-600 text-sm mb-2 truncate" title={manga.author}>
          by {manga.author}
        </p>

        {/* Genres */}
        {manga.genres && manga.genres.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-3">
            {manga.genres.slice(0, 3).map((genre, index) => (
              <span
                key={index}
                className="px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full"
              >
                {genre}
              </span>
            ))}
            {manga.genres.length > 3 && (
              <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                +{manga.genres.length - 3}
              </span>
            )}
          </div>
        )}

        {/* Meta Info */}
        <div className="flex items-center justify-between text-sm text-gray-500">
          <div className="flex items-center gap-1">
            <Book className="w-4 h-4" />
            <span>{manga.total_chapters || 0} chapters</span>
          </div>
          
          {manga.publication_year && (
            <div className="flex items-center gap-1">
              <Calendar className="w-4 h-4" />
              <span>{manga.publication_year}</span>
            </div>
          )}
        </div>

        {/* Status Badge */}
        {manga.status && (
          <div className="mt-3">
            <span
              className={`inline-block px-3 py-1 text-xs font-semibold rounded-full ${
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
      </div>
    </motion.div>
  );
};

export default MangaCard;
