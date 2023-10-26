import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const DisconnectButton = () => {
  const navigate = useNavigate();
  const [error, setError] = useState(null);

  const handleDisconnect = () => {
    setError(null);

    try {
      fetch('/api/oauth/logout', { method: 'POST', credentials: 'include' });

      document.cookie = "gcd_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;"

      console.log("document cookie: " + document.cookie)

      navigate('/');
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div>
      <button onClick={handleDisconnect}>Disconnect</button>
      {error && <p style={{ color: 'red' }}>{error}</p>}
    </div>
  );
};

export default DisconnectButton;