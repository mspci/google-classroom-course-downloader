import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const DisconnectButton = () => {
  const navigate = useNavigate();
  const [error, setError] = useState(null);

  const handleDisconnect = async () => {
    try {
      await fetch('http://localhost:8080/oauth/logout', { method: 'POST', credentials: 'include' });
      document.cookie = 'GCD_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
      navigate('/');
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <button onClick={handleDisconnect}>Disconnect</button>
    </>
  );
};

export default DisconnectButton;
