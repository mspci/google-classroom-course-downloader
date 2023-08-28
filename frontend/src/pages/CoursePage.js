import React, { useState } from 'react';
import CourseSelection from '../components/CourseSelection';
import CourseDownload from '../components/CourseDownload';

const CoursePage = () => {
    const [selectedCoursesIDs, setSelectedCoursesIDs] = useState([]);

    return (
        <div>
            <h1>Google Classroom Course Downloader</h1>
            <CourseSelection
                selectedCoursesIDs={selectedCoursesIDs}
                onCourseSelection={setSelectedCoursesIDs}
            />
            <CourseDownload selectedCoursesIDs={selectedCoursesIDs} />
        </div>
    );
};

export default CoursePage;
