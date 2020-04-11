import React from 'react';
export default function useLocalStorage(key: string, initialValue: string): [string | null, (value: string) => void] {
  const [storedValue, setStoredValue] = React.useState(() => {
    try {
      const item = window.localStorage.getItem(key);
      if (item === null) {
        window.localStorage.setItem(key, initialValue);
        return initialValue;
      }
      return item;
    } catch (error) {
      console.log(error);
      return null;
    }
  });

  const setValue = (value: string) => {
    try {
      setStoredValue(value);
      window.localStorage.setItem(key, value);
    } catch (error) {
      console.log(error);
    }
  };

  return [storedValue, setValue];
}
