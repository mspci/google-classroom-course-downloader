import React from 'react';
import CoursePage from './pages/CoursePage';
import ErrorPage from "./pages/ErrorPage";
import Login from './components/Login';
import {
  createBrowserRouter,
  RouterProvider,
} from "react-router-dom";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Login />,
    errorElement: <ErrorPage />,
  },
  {
    path: "/courses",
    element: <CoursePage />,
    errorElement: <ErrorPage />,
  },
]);

const App = () => {
  return (
    <RouterProvider router={router} />
  );
};

export default App;
