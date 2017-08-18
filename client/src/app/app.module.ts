import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';

import { AppComponent } from './app.component';
import { AppRoutingModule } from './app-routing.module';
import { EntityComponent } from './entity/entity.component';
import { EntityListComponent } from './entity/entity-list.component';
import { EntityNewComponent } from './entity/entity-new.component';
import { TopComponent } from './top/top.component';

@NgModule({
  declarations: [
    AppComponent,
    EntityComponent,
    EntityListComponent,
    EntityNewComponent,
    TopComponent,
  ],
  imports: [
    BrowserModule,
    ReactiveFormsModule,
    AppRoutingModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
