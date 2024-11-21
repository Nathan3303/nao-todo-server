import mongoose from 'mongoose';

const connectAndRun = async (fn: () => Promise<void>) => {
    try {
        await mongoose.connect('mongodb://172.17.0.5:27017/naotodo');
        await fn();
        await new Promise(resolve => setTimeout(resolve, 512));
    } catch (error) {
        console.error('[connect-and-run]', error);
    } finally {
        await mongoose.disconnect();
    }
};

export default connectAndRun;
