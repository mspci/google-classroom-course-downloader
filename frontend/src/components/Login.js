import React, { useState, useEffect } from 'react';

const Login = () => {
    const [loginError, setLoginError] = useState(null);

    useEffect(() => {
        login();
    }, []);

    const login = async () => {
        try {
            setLoginError(null);

            const response = await fetch('http://localhost:8080/', {
                method: 'GET',
                credentials: 'include',
            });

            if (response.status === 401) {
                const authResponse = await fetch('http://localhost:8080/oauth/url');
                const authData = await authResponse.json();
                window.location.href = authData.url;
            } else {
                window.location.href = '/courses';
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
