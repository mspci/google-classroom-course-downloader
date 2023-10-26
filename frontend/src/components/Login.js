import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const Login = () => {
    const [loginError, setLoginError] = useState(null);
    const navigate = useNavigate();

    useEffect(() => {
        login();
    }, []);

    const login = async () => {
        try {
            setLoginError(null);

            const response = await fetch('/api/', {
                method: 'GET',
                credentials: 'include',
            });

            if (response.status === 401) {
                const authResponse = await fetch('/api/oauth/url');
                const authData = await authResponse.json();
                window.location.href = authData.url
            } else {
                navigate('/courses')
            }
        } catch (error) {
            console.error('Error initiating login:', error);
            setLoginError('Failed to initiate login. Please try again later. ' + error);
        }
    };

    return (
        <div>
            <h1>Welcome to Google Classroom Downloader</h1>
            <p>Redirecting to login...</p>
            {loginError && <p style={{ color: 'red' }}>{loginError}</p>}
        </div>
    );
};

export default Login;
