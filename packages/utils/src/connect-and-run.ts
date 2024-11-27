import mongoose from 'mongoose';

const mongodbUrl = `mongodb://${PROD ? '172.18.0.3' : 'localhost'}:27017/naotodo`;

const connectAndRun = async (fn: () => Promise<void>) => {
    await mongoose.connect(mongodbUrl);
    await fn();
};

export default connectAndRun;
