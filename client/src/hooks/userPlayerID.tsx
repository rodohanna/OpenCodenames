import { v4 as uuidv4 } from 'uuid';
import useLocalStorage from './useLocalStorage';

export default function usePlayerID(): string | null {
  const [playerID] = useLocalStorage('playerID', uuidv4());
  return playerID;
}
