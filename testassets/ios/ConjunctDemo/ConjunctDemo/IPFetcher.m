//
//  IPFetcher.m
//  ConjunctDemo
//
//  Created by afjoseph on 29.09.23.
//

#import <Foundation/Foundation.h>
#import "IPFetcher.h"

@implementation IPFetcher

- (void)fetchIPAddressWithCompletion:(void (^)(NSString *ipAddress, NSError *error))completion {
    NSURL *url = [NSURL URLWithString:@"http://icanhazip.com"];
    NSURLSessionDataTask *task = [[NSURLSession sharedSession] dataTaskWithURL:url completionHandler:^(NSData * _Nullable data, NSURLResponse * _Nullable response, NSError * _Nullable error) {
        if (error) {
            if (completion) {
                completion(nil, error);
            }
            return;
        }

        if (data) {
            NSString *ipAddress = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
            ipAddress = [ipAddress stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceAndNewlineCharacterSet]]; // Remove new line and whitespace
            if (completion) {
                completion(ipAddress, nil);
            }
        }
    }];
    
    [task resume];
}

@end
