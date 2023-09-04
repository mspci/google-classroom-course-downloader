import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom'

const CourseDownload = ({ selectedCoursesIDs }) => {
    const [isDownloading, setIsDownloading] = useState(false);
    const navigate = useNavigate();

    const handleDownload = async () => {
        try {
            setIsDownloading(true);
            const response = await fetch('/api/courses/download', {
                credentials: 'include',
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ selectedCoursesIDs }),
            });

            if (response.status === 401) {
                navigate('/');
            } else {
                console.log('Download initiated successfully.');
                generateDownloadLink();
            }
        } catch (error) {
            console.error('Error sending download request:', error);
        } finally {
            setIsDownloading(false); // Download process completed
        }
    };

    const generateDownloadLink = () => {
        // const downloadLink = document.createElement('a');
        // downloadLink.href = '/api/courses/serve';
        // downloadLink.target = '_blank'; // Open in a new tab
        // downloadLink.download = 'downloaded_courses.zip'; // Specify the download file name
        // downloadLink.click();
    };

    return (
        <div>
            <h2>Download Selected Courses</h2>
            <button onClick={handleDownload} disabled={isDownloading || selectedCoursesIDs.length === 0}>
                {isDownloading ? 'Downloading...' : 'Download'}
            </button>
        </div>
    );
};

export default CourseDownload;
