import React, { useState } from 'react';
import { X } from 'lucide-react';

const AddToLibraryModal = ({ isOpen, onClose, manga, onAdd, onRemove, currentStatus }) => {
  const [selectedStatus, setSelectedStatus] = useState(currentStatus || 'reading');
  const [isLoading, setIsLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);

  // Update selected status when modal opens with current status
  React.useEffect(() => {
    if (isOpen) {
      setSelectedStatus(currentStatus || 'reading');
    }
  }, [isOpen, currentStatus]);

  const statusOptions = [
    { value: 'none', label: 'None' },
    { value: 'reading', label: 'Reading' },
    { value: 'on_hold', label: 'On Hold' },
    { value: 'dropped', label: 'Dropped' },
    { value: 'plan_to_read', label: 'Plan to Read' },
    { value: 'completed', label: 'Completed' },
    { value: 're_reading', label: 'Re-Reading' }
  ];

  const handleAdd = async () => {
    setIsLoading(true);
    try {
      if (selectedStatus === 'none') {
        // Remove from library if None is selected
        await onRemove();
      } else {
        // Add or update status
        await onAdd(selectedStatus);
      }
      onClose();
    } catch (error) {
      console.error('Failed to update library:', error);
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen || !manga) return null;

  const selectedLabel = statusOptions.find(opt => opt.value === selectedStatus)?.label || 'Reading';

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-background-dark rounded-xl max-w-2xl w-full shadow-xl border border-gray-200 dark:border-gray-800">
        {/* Header */}
        <div className="flex items-center justify-between p-8 border-b border-gray-200 dark:border-gray-800">
          <h2 className="text-3xl font-bold text-text-light dark:text-text-dark">
            Add To Library
          </h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            disabled={isLoading}
          >
            <X size={28} />
          </button>
        </div>

        {/* Content */}
        <div className="p-8">
          {/* Manga Info and Status Selector Side by Side */}
          <div className="flex gap-8">
            {/* Manga Cover */}
            <div className="flex-shrink-0">
              <img
                src={manga.cover_url || manga.main_picture?.large}
                alt={manga.title}
                className="w-56 h-80 object-cover rounded-lg shadow-md"
              />
            </div>

            {/* Right Side - Title and Dropdown */}
            <div className="flex-1 flex flex-col">
              <h3 className="font-bold text-2xl text-text-light dark:text-text-dark mb-3">
                {manga.title}
              </h3>

              {/* Reading Status Selector */}
              <div className="flex-1">
                <label className="block text-lg font-semibold text-text-light dark:text-text-dark mb-4">
                  Reading Status
                </label>
                <div className="relative">
                  <button
                    onClick={() => setShowDropdown(!showDropdown)}
                    className="w-full px-5 py-4 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-text-light dark:text-text-dark text-left flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors focus:outline-none focus:ring-2 focus:ring-primary text-lg"
                    disabled={isLoading}
                  >
                    <span>{selectedLabel}</span>
                    <svg className="w-6 h-6 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                  
                  {/* Dropdown Menu */}
                  {showDropdown && (
                    <div className="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 overflow-hidden max-h-80 overflow-y-auto">
                      {statusOptions.map(option => (
                        <button
                          key={option.value}
                          onClick={() => {
                            setSelectedStatus(option.value);
                            setShowDropdown(false);
                          }}
                          className={`w-full px-5 py-4 text-left transition-colors text-lg ${
                            selectedStatus === option.value
                              ? 'bg-primary text-white'
                              : 'text-text-light dark:text-text-dark hover:bg-gray-100 dark:hover:bg-gray-700'
                          }`}
                          disabled={isLoading}
                        >
                          {option.label}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex gap-4 p-8 border-t border-gray-200 dark:border-gray-800">
          <button
            onClick={onClose}
            className="flex-1 px-8 py-4 bg-gray-500 text-white rounded-lg hover:bg-gray-600 transition-colors font-semibold text-lg disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isLoading}
          >
            Cancel
          </button>
          <button
            onClick={handleAdd}
            className="flex-1 px-8 py-4 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors font-semibold text-lg disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isLoading}
          >
            {isLoading ? (selectedStatus === 'none' ? 'Removing...' : 'Updating...') : (selectedStatus === 'none' ? 'Remove' : (currentStatus ? 'Update' : 'Add'))}
          </button>
        </div>
      </div>
    </div>
  );
};

export default AddToLibraryModal;
