//
//  IPFetcher.h
//  ConjunctDemo
//
//  Created by afjoseph on 29.09.23.
//

#ifndef IPFetcher_h
#define IPFetcher_h

@interface IPFetcher : NSObject

- (void)fetchIPAddressWithCompletion:(void (^)(NSString *ipAddress, NSError *error))completion;

@end

#endif /* IPFetcher_h */
