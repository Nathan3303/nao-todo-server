import mongoose from 'mongoose';

const connectAndRun = async (fn: () => Promise<void>) => {
    try {
        await mongoose.connect(
            'mongodb://localhost:27017/test'
        );
        await fn();
    } catch (error) {
        console.error('[connect-and-run]', error);
    } finally {
        await mongoose.disconnect();
    }
};

export default connectAndRun;