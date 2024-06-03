import { initializeApp } from 'firebase/app';
import { getMessaging } from 'firebase/messaging';
import { firebaseConfig } from './config';

initializeApp(firebaseConfig);

export default getMessaging();