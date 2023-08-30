import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const CourseSelection = ({ selectedCoursesIDs, onCourseSelection }) => {
    const [courses, setCourses] = useState([]);
    const [fetchingStatus, setFetchingStatus] = useState('loading'); // 'loading', 'success', 'error'
    const navigate = useNavigate();

    useEffect(() => {
        fetchCourses();
    }, []);

    const fetchCourses = async () => {
        try {
            const response = await fetch('http://localhost:8080/courses/list', {
                credentials: 'include'
            });
            if (response.status === 401) {
                setFetchingStatus('error');
                console.error('Error fetching courses:', response.status);
                navigate('/')
            }
            const data = await response.json();
            setCourses(data);
            setFetchingStatus('success');
        } catch (error) {
            setFetchingStatus('error');
            console.error('Error fetching courses:', error);
        }
    };

    const handleCourseSelection = (courseId) => {
        const updatedSelectedCoursesIDs = selectedCoursesIDs.includes(courseId)
            ? selectedCoursesIDs.filter(id => id !== courseId)
            : [...selectedCoursesIDs, courseId];

        onCourseSelection(updatedSelectedCoursesIDs);
    };

    return (
        <div>
            {fetchingStatus === 'error' && <p style={{ color: 'red' }}>Error fetching courses. Please try again later.</p>}
            {fetchingStatus === 'loading' && <p style={{ color: 'green' }}>Loading courses...</p>}
            {fetchingStatus === 'success' && courses && courses.length > 0 && (
                <div>
                    <h2>Select Courses</h2>
                    <ul>
                        {courses.map(course => (
                            <li key={course.id}>
                                <label>
                                    <input
                                        type="checkbox"
                                        checked={selectedCoursesIDs.includes(course.id)}
                                        onChange={() => handleCourseSelection(course.id)}
                                    />
                                    {course.name}
                                </label>
                            </li>
                        ))}
                    </ul>
                </div>
            )}

        </div>
    );
};

export default CourseSelection;
