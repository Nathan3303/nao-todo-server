import mongoose from 'mongoose';

const mongodbUrl = `mongodb://${PROD ? '172.18.0.3' : 'localhost'}:27017/naotodo`;

const connectAndRun = async (fn: () => Promise<void>) => {
    try {
        await mongoose.connect(mongodbUrl);
        await fn();
        await new Promise(resolve => setTimeout(resolve, 512));
    } catch (error) {
        console.error('[connect-and-run]', error);
    } finally {
        await mongoose.disconnect();
    }
};

export default connectAndRun;
